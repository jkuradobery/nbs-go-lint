package separator_analyzer

import (
	"go/ast"
	"go/token"
	"log"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// between comment and separator can also be one line
// If comment contains several "/" it should be treated as a separator for which
// length != 80 is not allowed
////////////////////////////////////////////////////////////////////////////////

const separator = "////////////////////////////////////////////////////////////////////////////////"

////////////////////////////////////////////////////////////////////////////////

func getOriginalCommentText(
	fset *token.FileSet,
	commentGroup *ast.CommentGroup,
	data []byte,
) string {

	if commentGroup == nil {
		return ""
	}

	var builder strings.Builder
	for _, comment := range commentGroup.List {
		// Get the position information
		startPos := fset.Position(comment.Slash)
		endPos := fset.Position(comment.End())

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
			}

			return nil, nil
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

type SeparatorAnalysis struct {
	pass                 *analysis.Pass
	fset                 *token.FileSet
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
		fset:                 pass.Fset,
		file:                 file,
		topLevelDeclarations: topLevelDeclarations,
		separators: slices.SortedFunc(
			slices.Values(
				Filter(
					file.Comments,
					func(group *ast.CommentGroup) bool {
						text := getOriginalCommentText(pass.Fset, group, data)
						return strings.Contains(text, separator)
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
			Category: "separator",
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
				Category: "separator",
				Message:  "Separator is not allowed to be before package declaration",
			})
		}
		return
	}

	lastImport := s.imports[len(s.imports)-1]
	if firstSeparator.Pos() < lastImport.End() {
		s.pass.Report(analysis.Diagnostic{
			Pos:      firstSeparator.Pos(),
			End:      firstSeparator.End(),
			Category: "separator",
			Message:  "Separator is not allowed to be before imports",
		})
	}
}

func (s *SeparatorAnalysis) ForbiddenMultilineComments() {
	for _, group := range s.separators {
		startLine := s.fset.Position(group.Pos()).Line
		endLine := s.fset.Position(group.End()).Line
		if endLine-startLine > 0 {
			s.pass.Report(analysis.Diagnostic{
				Pos:      group.Pos(),
				End:      group.End(),
				Category: "separator",
				Message:  "Separator is not allowed to be a part of multiline comment",
			})
		}
	}
}

func (s *SeparatorAnalysis) ForbiddenSeparatorOverCode() {
	if len(s.separators) == 0 {
		return
	}

	for _, separator := range s.separators {
		// MxN instead of M + N compexity for code simplicity
		for _, declaration := range s.topLevelDeclarations {
			if s.nodesOverlap(separator, declaration) {
				s.pass.Report(analysis.Diagnostic{
					Pos:      separator.Pos(),
					End:      separator.End(),
					Category: "separator",
					Message:  "Separator is not allowed to be over code",
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
		lineIndex := s.fset.Position(separator.Pos()).Line - 1
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
			Category: "separator",
			Message:  "Each separator should be surrounded by exactly one empty line",
		})
	}
}

func (s *SeparatorAnalysis) NoDeclarationsBetweenTwoSeparators() {

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
	return s.fset.Position(pos)
}
