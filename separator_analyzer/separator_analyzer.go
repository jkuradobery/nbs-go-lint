package separator_analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"slices"
	"strings"

	set "github.com/deckarep/golang-set/v2"
	"golang.org/x/tools/go/analysis"
)

// between comment and separator can also be one line
// If comment contains several "/" it should be treated as a separator for which
// length != 80 is not allowed

////////////////////////////////////////////////////////////////////////////////

const Separator = "////////////////////////////////////////////////////////////////////////////////"
const emptyReceiver = "emptyReceiver"
const analyzerCategory = "separator"
const MixingPublicAndPrivate = "Mixing public and private methods in the same group is not allowed"
const MixingTestingAndCode = "Mixing testing and code methods in the same group is not allowed"
const MixingMethodsWithIncorrectReceiverFormat = "Mixing methods with different receivers in the same group is not allowed %s"
const SingleInterfaceOrStructMessage = "Only one interface or struct declaration is allowed between separators"

////////////////////////////////////////////////////////////////////////////////

func getOriginalCommentText(
	fileset *token.FileSet,
	commentGroup *ast.CommentGroup,
	data []byte,
) string {

	if commentGroup == nil {
		return ""
	}

	var builder strings.Builder
	for _, comment := range commentGroup.List {
		// Get the position information
		startPos := fileset.Position(comment.Slash)
		endPos := fileset.Position(comment.End())

		// Extract the original text
		originalText := data[startPos.Offset:endPos.Offset]
		builder.Write(originalText)
	}

	return builder.String()
}

////////////////////////////////////////////////////////////////////////////////

func Filter[T any](data []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(data))
	for _, item := range data {
		if !predicate(item) {
			continue
		}

		result = append(result, item)
	}

	return result
}

////////////////////////////////////////////////////////////////////////////////

type functionDeclarationType struct {
	receiver  string
	isPublic  bool
	isTesting bool
}

func (f *functionDeclarationType) String() string {
	return fmt.Sprintf(
		"receiver: %s, isPublic: %v, isTesting: %v",
		f.receiver,
		f.isPublic,
		f.isTesting,
	)
}

////////////////////////////////////////////////////////////////////////////////

type functionDeclarationStorage struct {
	declarationsForType map[string][]*ast.FuncDecl
	declarationTypes    []functionDeclarationType
}

func newFunctionDeclarationStorage() *functionDeclarationStorage {
	return &functionDeclarationStorage{
		declarationsForType: make(map[string][]*ast.FuncDecl),
		declarationTypes:    make([]functionDeclarationType, 0),
	}
}

func (f *functionDeclarationStorage) Add(decl *ast.FuncDecl) {
	declarationType := functionDeclarationType{
		receiver: emptyReceiver,
	}
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		// If the function has a receiver, we consider it as a method
		// and use the receiver type as part of the key.
		if ident, ok := decl.Recv.List[0].Type.(*ast.Ident); ok {
			declarationType.receiver = ident.Name
		} else if starExpr, ok := decl.Recv.List[0].Type.(*ast.StarExpr); ok {
			if ident, ok := starExpr.X.(*ast.Ident); ok {
				declarationType.receiver = ident.Name
			}
		}
	}

	if decl.Name.IsExported() {
		declarationType.isPublic = true
	}

	if strings.HasPrefix(decl.Name.Name, "Test") {
		declarationType.isTesting = true
	}

	key := declarationType.String()
	if _, exists := f.declarationsForType[key]; !exists {
		f.declarationsForType[key] = make([]*ast.FuncDecl, 0)
		f.declarationTypes = append(f.declarationTypes, declarationType)
	}
	f.declarationsForType[key] = append(f.declarationsForType[key], decl)
}
func (f *functionDeclarationStorage) MixedTestingAndCode() []*ast.FuncDecl {
	if len(f.declarationTypes) == 0 {
		return []*ast.FuncDecl{}
	}

	result := make([]*ast.FuncDecl, 0)
	first := f.declarationTypes[0]
	for _, declType := range f.declarationTypes {
		if declType.isTesting != first.isTesting {
			if len(result) > 0 {
				continue
			}

			result = append(
				result,
				f.declarationsForType[declType.String()][0],
				f.declarationsForType[first.String()][0],
			)
		}
	}

	return result
}

func (f *functionDeclarationStorage) MixedPublicAndPrivate() []*ast.FuncDecl {
	if len(f.declarationTypes) == 0 {
		return []*ast.FuncDecl{}
	}

	result := make([]*ast.FuncDecl, 0)
	first := f.declarationTypes[0]
	for _, declType := range f.declarationTypes {
		if declType.isPublic != first.isPublic {

			if len(result) > 0 {
				result = append(
					result,
					f.declarationsForType[first.String()]...,
				)
			}
			result = append(
				result,
				f.declarationsForType[declType.String()]...,
			)
		}
	}

	return result
}

func (f *functionDeclarationStorage) DeclarationsByReceiver() map[string][]*ast.FuncDecl {

	result := make(map[string][]*ast.FuncDecl)
	for _, declType := range f.declarationTypes {
		result[declType.receiver] = f.declarationsForType[declType.String()]
	}

	return result
}

////////////////////////////////////////////////////////////////////////////////

func functionReturnsStruct(decl *ast.FuncDecl, structName string) bool {
	if decl.Type == nil || decl.Type.Results == nil {
		return false
	}

	for _, result := range decl.Type.Results.List {
		switch t := result.Type.(type) {
		case *ast.Ident:
			if t.Name == structName {
				return true
			}
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok && ident.Name == structName {
				if ident.Name == structName {
					return true
				}
			}
		}
	}

	// Find all return expressions in the child expressions
	variableTypesByName := getVariableTypesByName(decl)

	for _, expr := range decl.Body.List {
		returnStmt, ok := expr.(*ast.ReturnStmt)
		if !ok {
			continue
		}

		// we do not deduce the type from something complicated,
		// constructors are to be made simple, either var or assignment
		// Also we do not care for returned types only for references to a type
		for _, returnExpression := range returnStmt.Results {
			// Check if one of results is a pointer to struct
			//
			if getReferencedStructType(returnExpression) == structName {
				return true
			}

			if resIdent, ok := returnExpression.(*ast.Ident); ok {
				if returnType, ok := variableTypesByName[resIdent.Name]; ok {
					if returnType == structName {
						return true
					}
				}
			}
		}
	}

	return false
}

func getReferencedStructType(decl ast.Expr) string {
	unaryExpr, ok := decl.(*ast.UnaryExpr)
	if !ok {
		return ""
	}

	if unaryExpr.Op != token.AND {
		return ""
	}

	x := unaryExpr.X
	compositeLit, ok := x.(*ast.CompositeLit)
	if !ok {
		return ""
	}

	ident, ok := compositeLit.Type.(*ast.Ident)
	if !ok {
		return ""
	}

	return ident.Name
}

func getVariableTypesByName(decl *ast.FuncDecl) map[string]string {
	result := make(map[string]string)
	for _, stmt := range decl.Body.List {
		if assignStmt, ok := stmt.(*ast.AssignStmt); ok {
			for index, lhs := range assignStmt.Lhs {
				ident, ok := lhs.(*ast.Ident)
				if !ok {
					continue
				}

				name := ident.Name
				if len(assignStmt.Rhs) <= index {
					// If the value was returned from a function
					// we do not want to deduce the type.
					// It is a constructor, it should construct.
					continue
				}
				leftSide := assignStmt.Rhs[index]
				structType := getReferencedStructType(leftSide)
				if structType != "" {
					result[name] = structType
				}
			}
		}
	}

	return result
}

////////////////////////////////////////////////////////////////////////////////

func SeparatorAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "SeparatorAnalyzer",
		Doc:  "Checks if 80 lines 'otbivka' separates logical entities",
		Run: func(pass *analysis.Pass) (any, error) {
			for _, file := range pass.Files {
				separatorAnalysis := NewSeparatorAnalysis(pass, file)
				separatorAnalysis.ForbiddenSeparatorAtTheEnd()
				separatorAnalysis.ForbiddenSeparatorBeforeImports()
				separatorAnalysis.ForbiddenMultilineComments()
				separatorAnalysis.ForbiddenSeparatorOverCode()
				separatorAnalysis.EmptyLinesAroundSeparator()
				separatorAnalysis.NoDeclarationsBetweenTwoSeparators()
				separatorAnalysis.CheckSeparatorAfterPackageForMissingImport()
				separatorAnalysis.CheckSeparatorGroupsCorrectEntities()
			}

			return nil, nil
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type SeparatorAnalysis struct {
	pass                 *analysis.Pass
	fileset              *token.FileSet
	file                 *ast.File
	topLevelDeclarations []ast.Decl
	imports              []*ast.ImportSpec
	separators           []*ast.CommentGroup
	data                 []byte
	lines                []string
}

func NewSeparatorAnalysis(
	pass *analysis.Pass,
	file *ast.File,
) SeparatorAnalysis {
	comparator := func(i ast.Node, j ast.Node) int {
		if i.Pos() > j.Pos() {
			return 1
		}

		if i.Pos() < j.Pos() {
			return -1
		}

		if i.End() > j.End() {
			return 1
		}

		if i.End() < j.End() {
			return -1
		}

		return 0
	}

	topLevelDeclarations := slices.SortedFunc(
		slices.Values(file.Decls),
		func(decl ast.Decl, decl2 ast.Decl) int {
			return comparator(decl, decl2)
		},
	)

	data, err := pass.ReadFile(pass.Fset.Position(file.Pos()).Filename)
	if err != nil {
		log.Fatalf(
			"Error reading file %v: %v",
			pass.Fset.Position(file.Pos()).Filename,
			err,
		)
	}

	lines := strings.Split(string(data), "\n")
	return SeparatorAnalysis{
		pass:                 pass,
		fileset:              pass.Fset,
		file:                 file,
		topLevelDeclarations: topLevelDeclarations,
		separators: slices.SortedFunc(
			slices.Values(
				Filter(
					file.Comments,
					func(group *ast.CommentGroup) bool {
						text := getOriginalCommentText(pass.Fset, group, data)
						return strings.Contains(text, Separator)
					},
				),
			),
			func(group *ast.CommentGroup, group2 *ast.CommentGroup) int {
				return comparator(group, group2)
			},
		),
		imports: slices.SortedFunc(
			slices.Values(file.Imports),
			func(spec *ast.ImportSpec, spec2 *ast.ImportSpec) int {
				return comparator(spec, spec2)
			},
		),
		data:  data,
		lines: lines,
	}
}

////////////////////////////////////////////////////////////////////////////////

func (s *SeparatorAnalysis) ForbiddenSeparatorAtTheEnd() {
	if len(s.separators) == 0 {
		// No separators found, should be error in another function
		return
	}

	if len(s.topLevelDeclarations) == 0 {
		// No top level declarations found, this is probably a constant
		return
	}

	lastSeparator := s.separators[len(s.separators)-1]
	lastDeclaration := s.topLevelDeclarations[len(s.topLevelDeclarations)-1]
	if lastSeparator.Pos() > lastDeclaration.End() {
		s.pass.Report(analysis.Diagnostic{
			Pos:      lastSeparator.Pos(),
			End:      lastSeparator.End(),
			Category: analyzerCategory,
			Message:  "Separators at the end of the file are not allowed",
		})
	}
}

func (s *SeparatorAnalysis) ForbiddenSeparatorBeforeImports() {
	if len(s.separators) == 0 {
		return
	}

	firstSeparator := s.separators[0]

	if len(s.imports) == 0 {
		if firstSeparator.End() <= s.file.Package {
			s.pass.Report(analysis.Diagnostic{
				Pos:      firstSeparator.Pos(),
				End:      firstSeparator.End(),
				Category: analyzerCategory,
				Message:  "Separator is not allowed before package declaration",
			})
		}
		return
	}

	lastImport := s.imports[len(s.imports)-1]
	if firstSeparator.Pos() < lastImport.End() {
		s.pass.Report(analysis.Diagnostic{
			Pos:      firstSeparator.Pos(),
			End:      firstSeparator.End(),
			Category: analyzerCategory,
			Message:  "Separator is not allowed before imports",
		})
	}
}

func (s *SeparatorAnalysis) ForbiddenMultilineComments() {
	for _, group := range s.separators {
		startLine := s.fileset.Position(group.Pos()).Line
		endLine := s.fileset.Position(group.End()).Line
		if endLine-startLine > 0 {
			s.pass.Report(analysis.Diagnostic{
				Pos:      group.Pos(),
				End:      group.End(),
				Category: analyzerCategory,
				Message:  "Separator is not allowed a part of multiline comment",
			})
		}
	}
}

func (s *SeparatorAnalysis) ForbiddenSeparatorOverCode() {
	if len(s.separators) == 0 {
		return
	}

	for _, separator := range s.separators {
		// MxN instead of M + N complexity for code simplicity
		for _, declaration := range s.topLevelDeclarations {
			if s.nodesOverlap(separator, declaration) {
				s.pass.Report(analysis.Diagnostic{
					Pos:      separator.Pos(),
					End:      separator.End(),
					Category: analyzerCategory,
					Message:  "Separator is not allowed over code",
				})
			}
		}
	}
}

func (s *SeparatorAnalysis) EmptyLinesAroundSeparator() {
	if len(s.separators) == 0 {
		return
	}

	if len(s.lines) == 0 {
		return
	}

	for _, separator := range s.separators {
		lineIndex := s.position(separator.Pos()).Line - 1
		emptyLinesBefore := 0
		emptyLinesAfter := 0
		for i := lineIndex + 1; i < len(s.lines); i++ {
			if s.lines[i] != "" {
				break
			}

			emptyLinesAfter++
		}

		for i := lineIndex - 1; i >= 0; i-- {
			if s.lines[i] != "" {
				break
			}

			emptyLinesBefore++
		}

		if emptyLinesBefore == 1 && emptyLinesAfter == 1 {
			continue
		}

		if emptyLinesAfter == 0 && lineIndex == len(s.lines)-1 {
			if emptyLinesBefore == 1 {
				continue
			}
		}

		if emptyLinesBefore == 0 && lineIndex == 0 {
			if emptyLinesAfter == 1 {
				continue
			}
		}

		s.pass.Report(analysis.Diagnostic{
			Pos:      separator.Pos(),
			End:      separator.End(),
			Category: analyzerCategory,
			Message:  "Each Separator should be surrounded by exactly one empty line",
		})
	}
}

func (s *SeparatorAnalysis) NoDeclarationsBetweenTwoSeparators() {
	if len(s.separators) < 2 {
		return
	}

	for i := 0; i < len(s.separators)-1; i++ {
		currentSeparator := s.separators[i]
		nextSeparator := s.separators[i+1]
		declarationFound := false
		for _, declaration := range s.topLevelDeclarations {
			if declaration.Pos() > currentSeparator.End() && declaration.End() < nextSeparator.Pos() {
				declarationFound = true
				break
			}
		}

		if !declarationFound {
			s.pass.Report(analysis.Diagnostic{
				Pos:      currentSeparator.End(),
				End:      nextSeparator.Pos(),
				Category: analyzerCategory,
				Message:  "Empty section detected: no declarations found between consecutive separators",
			})
		}
	}
}

func (s *SeparatorAnalysis) CheckSeparatorAfterPackageForMissingImport() {
	if len(s.imports) > 0 {
		return
	}

	const message = "Missing Separator after package " +
		"declaration when no imports present"
	if len(s.separators) == 0 {
		s.pass.Report(analysis.Diagnostic{
			Pos:      s.file.Package,
			End:      s.file.Package,
			Category: analyzerCategory,
			Message:  message,
		})
		return
	}

	firstSeparator := s.separators[0]
	packageLine := s.position(s.file.Package).Line
	separatorLine := s.position(firstSeparator.Pos()).Line
	if separatorLine != packageLine+2 {
		s.pass.Report(
			analysis.Diagnostic{
				Pos:      s.file.Package,
				End:      s.file.Package,
				Category: analyzerCategory,
				Message:  message,
			},
		)
	}
}

func (s *SeparatorAnalysis) CheckSeparatorGroupsCorrectEntities() {
	// Type alias groups, var groups, const groups, import groups should be separated
	// by a single separator.
	// Exactly one interface can be declared between two separators.
	// Test functions should be separated by a single separator.
	// Exactly one struct and its constructor and its private methods should be separated.
	// Mixing of public and private methods in the same group is not allowed.
	for _, bucket := range s.collectDeclarationsByBuckets() {
		if len(bucket) == 0 {
			continue
		}

		declarationsByTypeWithinBucket := make(map[token.Token][]ast.Decl)
		for _, decl := range bucket {
			var tok token.Token
			switch d := decl.(type) {
			case *ast.GenDecl:
				tok = s.getTokenForTopLevelDecl(d)
			case *ast.FuncDecl:
				tok = token.FUNC
			default:
				message := "Unknown declaration type found in bucket, might be a bug"
				s.pass.Report(
					analysis.Diagnostic{
						Pos:      decl.Pos(),
						End:      decl.End(),
						Category: analyzerCategory,
						Message:  message,
					},
				)
				continue
			}

			if _, ok := declarationsByTypeWithinBucket[tok]; !ok {
				declarationsByTypeWithinBucket[tok] = make([]ast.Decl, 0)
			}

			declarationsByTypeWithinBucket[tok] = append(
				declarationsByTypeWithinBucket[tok],
				decl,
			)
			s.processDeclarationsWithinBucket(declarationsByTypeWithinBucket)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

func (s *SeparatorAnalysis) nodesOverlap(node ast.Node, node2 ast.Node) bool {
	if s.position(node.Pos()).Line > s.position(node2.End()).Line {
		return false
	} else if s.position(node.End()).Line < s.position(node2.Pos()).Line {
		return false
	}

	return true
}

func (s *SeparatorAnalysis) position(pos token.Pos) token.Position {
	return s.fileset.Position(pos)
}

func (s *SeparatorAnalysis) collectDeclarationsByBuckets() map[int][]ast.Decl {
	// declarations between j - 1 and j separators go to bucket j
	buckets := make(map[int][]ast.Decl)
	if len(s.topLevelDeclarations) == 0 {
		return buckets
	}

	if len(s.separators) == 0 {
		buckets[0] = s.topLevelDeclarations
		return buckets
	}

	buckets[0] = make([]ast.Decl, 0, len(s.topLevelDeclarations))
	firstSeparator := s.separators[0]
	for _, declaration := range s.topLevelDeclarations {
		if declaration.Pos() < firstSeparator.Pos() {
			continue
		}
	}

	for i := 0; i < len(s.separators)-1; i++ {

		currentSeparator := s.separators[i]
		nextSeparator := s.separators[i+1]
		buckets[i+1] = make([]ast.Decl, 0, len(s.topLevelDeclarations))
		for _, declaration := range s.topLevelDeclarations {
			if declaration.Pos() <= currentSeparator.End() {
				continue
			}

			if declaration.End() >= nextSeparator.Pos() {
				continue
			}

			buckets[i+1] = append(buckets[i+1], declaration)
		}
	}

	lastSeparator := s.separators[len(s.separators)-1]
	buckets[len(s.separators)] = make([]ast.Decl, 0, len(s.topLevelDeclarations))
	for _, declaration := range s.topLevelDeclarations {
		if declaration.Pos() > lastSeparator.End() {
			buckets[len(s.separators)] = append(buckets[len(s.separators)], declaration)
		}
	}

	return buckets
}

func (s *SeparatorAnalysis) processDeclarationsWithinBucket(
	declarationsByTypeWithinBucket map[token.Token][]ast.Decl,
) {
	singleDeclarationTypeRequired := map[token.Token]struct{}{
		token.TYPE:   {},
		token.VAR:    {},
		token.CONST:  {},
		token.IMPORT: {},
	}
	if len(declarationsByTypeWithinBucket) == 0 {
		return
	}

	storage := newFunctionDeclarationStorage()
	if functionDecls, ok := declarationsByTypeWithinBucket[token.FUNC]; ok {
		for _, decl := range functionDecls {
			if decl, ok := decl.(*ast.FuncDecl); ok {
				storage.Add(decl)
			}
		}
	}

	// Check for single interface or struct declaration
	for _, tokenType := range []token.Token{token.INTERFACE, token.STRUCT} {
		if decls, ok := declarationsByTypeWithinBucket[tokenType]; ok {
			if len(decls) > 1 {
				s.reportMultipleInterfacesOrStructs(decls)
				return
			}
		}
	}

	if len(declarationsByTypeWithinBucket) == 1 {
		for declType := range declarationsByTypeWithinBucket {
			if _, ok := singleDeclarationTypeRequired[declType]; ok {
				return
			}

			if declType == token.INTERFACE {
				return
			}

			if declType == token.FUNC {
				s.reportIncorrectFunctionsSeparation(storage)
				return
			}
		}
	}

	if len(declarationsByTypeWithinBucket) > 2 {
		s.reportVariousTypesBetweenSeparators(declarationsByTypeWithinBucket)
		return
	}

	allowedTogether := set.NewSet(token.FUNC, token.STRUCT)
	keys := set.NewSetFromMapKeys(declarationsByTypeWithinBucket)
	if !keys.IsSubset(allowedTogether) {
		s.reportVariousTypesBetweenSeparators(declarationsByTypeWithinBucket)
		return
	}

	s.reportMixingTestsWithCode(storage)
	// TODO should we report mixing public and private methods here?
	structName := ""
	if structDecls, ok := declarationsByTypeWithinBucket[token.STRUCT]; ok {
		spec := structDecls[0].(*ast.GenDecl).Specs[0]
		structName = spec.(*ast.TypeSpec).Name.Name
	}

	for receiver, declarations := range storage.DeclarationsByReceiver() {
		if receiver == emptyReceiver {
			for _, decl := range declarations {
				if !functionReturnsStruct(decl, structName) {
					s.pass.Report(
						analysis.Diagnostic{
							Pos:      decl.Pos(),
							End:      decl.End(),
							Category: analyzerCategory,
							Message: fmt.Sprintf(
								"Function which is not a constructor for struct '%s' is not allowed in the same group as struct '%s'",
								decl.Name.Name,
								structName,
							),
						},
					)
				}
			}
		} else if receiver != structName {
			for _, decl := range declarations {
				s.pass.Report(
					analysis.Diagnostic{
						Pos:      decl.Pos(),
						End:      decl.End(),
						Category: analyzerCategory,
						Message: fmt.Sprintf(
							"Method with receiver '%s' is not allowed in the same group as struct '%s'",
							receiver,
							structName,
						),
					},
				)
			}
		}
	}
}

func (s *SeparatorAnalysis) reportMultipleInterfacesOrStructs(
	decls []ast.Decl,
) {
	s.pass.Report(analysis.Diagnostic{
		Pos:      decls[0].Pos(),
		End:      decls[len(decls)-1].End(),
		Category: analyzerCategory,
		Message:  SingleInterfaceOrStructMessage,
	})
}

func (s *SeparatorAnalysis) reportVariousTypesBetweenSeparators(
	declarationsByTypeWithinBucket map[token.Token][]ast.Decl,
) {

	declTypeList := make([]string, 0, len(declarationsByTypeWithinBucket))
	for declType := range declarationsByTypeWithinBucket {
		declTypeList = append(declTypeList, declType.String())
	}
	for _, declarations := range declarationsByTypeWithinBucket {
		for _, decl := range declarations {
			s.pass.Report(analysis.Diagnostic{
				Pos:      decl.Pos(),
				End:      decl.End(),
				Category: analyzerCategory,
				Message: fmt.Sprintf(
					"Forbidden declarations within the same group: %s",
					strings.Join(declTypeList, ", "),
				),
			})
		}
	}
}

func (s *SeparatorAnalysis) reportIncorrectFunctionsSeparation(
	storage *functionDeclarationStorage,
) {
	s.reportMixingPrivateAndPublic(storage)
	s.reportMixingTestsWithCode(storage)
	s.reportMixingDifferentReceiver(storage)
}

func (s *SeparatorAnalysis) reportMixingDifferentReceiver(
	storage *functionDeclarationStorage,
) {
	if receivers := storage.DeclarationsByReceiver(); len(receivers) > 1 {
		receiversList := make([]string, 0, len(receivers))
		for receiver := range receivers {
			receiversList = append(
				receiversList,
				fmt.Sprintf("'%s'", receiver),
			)
		}
		message := fmt.Sprintf(
			MixingMethodsWithIncorrectReceiverFormat,
			strings.Join(receiversList, ", "),
		)
		for _, decls := range receivers {
			for _, decl := range decls {
				s.pass.Report(
					analysis.Diagnostic{
						Pos:      decl.Pos(),
						End:      decl.End(),
						Category: analyzerCategory,
						Message:  message,
					},
				)
			}
		}
	}
}

func (s *SeparatorAnalysis) reportMixingTestsWithCode(storage *functionDeclarationStorage) {
	if decls := storage.MixedTestingAndCode(); len(decls) > 0 {
		for _, decl := range decls {
			s.pass.Report(analysis.Diagnostic{
				Pos:      decl.Pos(),
				End:      decl.End(),
				Category: analyzerCategory,
				Message:  MixingTestingAndCode,
			})
		}
	}
}

func (s *SeparatorAnalysis) reportMixingPrivateAndPublic(storage *functionDeclarationStorage) {
	if decls := storage.MixedPublicAndPrivate(); len(decls) > 0 {
		for _, decl := range decls {
			s.pass.Report(analysis.Diagnostic{
				Pos:      decl.Pos(),
				End:      decl.End(),
				Category: analyzerCategory,
				Message:  MixingPublicAndPrivate,
			})
		}
	}
}

func (s *SeparatorAnalysis) getTokenForTopLevelDecl(
	decl *ast.GenDecl,
) token.Token {
	// GenDecl can be a type, var, const or import declaration
	// for type declaration we want to return if it is an interface or struct
	// or a type alias. Multiple specs declaration is not allowed.
	// The following code is forbidden:
	// type(
	// 	a struct{
	//     b int
	// 	}
	// 	c interface{
	//     V() int
	// 	}
	// )
	if decl.Tok != token.TYPE {
		return decl.Tok
	}

	if len(decl.Specs) != 1 {
		const message = "Type declaration should have exactly one spec"
		s.pass.Report(
			analysis.Diagnostic{
				Pos:      decl.Pos(),
				End:      decl.End(),
				Category: analyzerCategory,
				Message:  message,
			},
		)
		return decl.Tok
	}

	if typeDecl, ok := decl.Specs[0].(*ast.TypeSpec); ok {
		switch typeDecl.Type.(type) {
		case *ast.InterfaceType:
			return token.INTERFACE
		case *ast.StructType:
			return token.STRUCT
		default:
			return token.TYPE
		}
	}

	return token.TYPE
}
