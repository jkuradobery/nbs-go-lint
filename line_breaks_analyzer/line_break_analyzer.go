package line_breaks_analyzer

import (
	set "github.com/deckarep/golang-set/v2"
	"go/ast"
	"log"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// //////////////////////////////////////////////////////////////////////////////

func LineBreakAfterRbracket() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "LineBreakAfterRbracket",
		Doc:  "Checks for line breaks after code block closures.",
		Run: func(pass *analysis.Pass) (any, error) {
			for _, file := range pass.Files {
				fset := pass.Fset
				data, err := pass.ReadFile(fset.Position(file.Pos()).Filename)
				if err != nil {
					log.Fatalf(
						"Error reading file %v: %v",
						fset.Position(file.Pos()).Filename,
						err,
					)
				}

				lines := strings.Split(string(data), "\n")
				functionBodyLbracketsByLine := set.NewSet[int]()
				ast.Inspect(file, func(n ast.Node) bool {
					if funcDecl, ok := n.(*ast.FuncDecl); ok {
						// We have a convention that line numbers start from 0,
						// and it should be maintained throughout the linter.
						line := fset.Position(funcDecl.Body.Lbrace).Line - 1
						functionBodyLbracketsByLine.Add(line)
					}

					if funcLit, ok := n.(*ast.FuncLit); ok {
						line := fset.Position(funcLit.Body.Lbrace).Line - 1
						functionBodyLbracketsByLine.Add(line)
					}

					return true
				})
				ast.Inspect(file, func(node ast.Node) bool {
					if blockStatement, ok := node.(*ast.BlockStmt); ok {
						checkNoLineBreakBeforeRbracket(
							pass,
							blockStatement,
							lines,
						)
						checkLineBreakAfterRbracket(
							pass,
							blockStatement,
							lines,
						)
					}

					if deferStatement, ok := node.(*ast.DeferStmt); ok {
						checkNoNewLineBeforeDeferStatement(
							pass,
							deferStatement,
							lines,
							functionBodyLbracketsByLine,
						)
					}
					return true
				})
			}

			return nil, nil
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

func checkNoLineBreakBeforeRbracket(
	pass *analysis.Pass,
	blockStatement *ast.BlockStmt,
	lines []string,
) {
	rbrace := blockStatement.Rbrace
	lbraceLine := pass.Fset.Position(blockStatement.Lbrace).Line
	rbraceLine := pass.Fset.Position(rbrace).Line
	if rbraceLine-lbraceLine < 2 {
		return
	}

	previousLineIndex := rbraceLine - 2
	// .Line indexing starts from 1,
	// so we need to subtract 2 to get the previous line
	if previousLineIndex < 0 || previousLineIndex >= len(lines) {
		return
	}

	previousLine := strings.TrimSpace(lines[previousLineIndex])
	if previousLine != "" {
		return
	}

	pass.Report(analysis.Diagnostic{
		Pos:      rbrace,
		End:      0,
		Category: "line_breaks",
		Message:  "Line break before closing } is not allowed.",
	})
}

func checkLineBreakAfterRbracket(
	pass *analysis.Pass,
	blockStatement *ast.BlockStmt,
	lines []string,
) {

	rbrace := blockStatement.Rbrace
	rbracePosition := pass.Fset.Position(
		rbrace,
	)
	nextLineIndex := rbracePosition.Line //  .Line indexing starts from 1
	if nextLineIndex >= len(lines) {
		return
	}
	nextLine := strings.TrimSpace(lines[nextLineIndex])
	if nextLine == "" {
		return
	}

	if strings.Contains(nextLine, "}") {
		return
	}

	currentLineIndex := rbracePosition.Line - 1
	if currentLineIndex >= 0 && currentLineIndex < len(lines) {
		afterBracket := strings.TrimSpace(
			lines[currentLineIndex][rbracePosition.Column:],
		)
		beforeComment := strings.Split(afterBracket, `//`)[0]
		beforeComment = regexp.MustCompile(`/\*.+\*/`).ReplaceAllString(
			beforeComment, "")
		if regexp.MustCompile(".*[,})].*").MatchString(beforeComment) {
			return
		}
	}

	if strings.HasPrefix(nextLine, "defer") {
		return
	}

	pass.Report(analysis.Diagnostic{
		Pos:      rbrace,
		End:      0,
		Category: "line_breaks",
		Message:  "Line break after closing } is required.",
	})
}

func checkNoNewLineBeforeDeferStatement(
	pass *analysis.Pass,
	deferStatement *ast.DeferStmt,
	lines []string,
	functionBodyLbracketsByLine set.Set[int],
) {
	deferStmtPos := deferStatement.Pos()
	previousLineIndex := pass.Fset.Position(deferStmtPos).Line - 2
	// .Line indexing starts from 1,
	// so we need to subtract 2 to get the previous line
	if previousLineIndex < 0 || previousLineIndex >= len(lines) {
		return
	}

	previousLine := strings.TrimSpace(lines[previousLineIndex])
	if previousLine != "" {
		return
	}

	if functionBodyLbracketsByLine.Contains(previousLineIndex - 1) {
		// If the previous line is a function body opening brace,
		// we allow the line break before the defer statement.
		return
	}
	pass.Report(analysis.Diagnostic{
		Pos:      deferStmtPos,
		End:      0,
		Category: "line_breaks",
		Message:  "Line break before 'defer' statement is not allowed.",
	})
}
