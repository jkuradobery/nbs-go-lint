package line_breaks_analyzer

import (
	"go/ast"
	"go/token"
	"log"
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
				rbrackets := make(map[token.Pos]int)
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

						if count, ok := rbrackets[blockStatement.Rbrace]; ok {
							rbrackets[blockStatement.Rbrace] = count + 1
						} else {
							rbrackets[blockStatement.Rbrace] = 1
						}
					}

					if deferStatement, ok := node.(*ast.DeferStmt); ok {
						checkNoNewLineBeforeDeferStatement(
							pass,
							deferStatement,
							lines,
						)
					}
					return true
				})

				for pos, count := range rbrackets {
					if count == 1 {
						continue
					}

					pass.Report(analysis.Diagnostic{
						Pos:      pos,
						End:      0,
						Category: "line_breaks",
						Message: "Multiple closing } brackets " +
							"on the same line are not allowed.",
					})
				}
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
	previousLineIndex := pass.Fset.Position(
		rbrace,
	).Line - 2
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
	nextLineIndex := pass.Fset.Position(
		rbrace,
	).Line //  .Line indexing starts from 1
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

	pass.Report(analysis.Diagnostic{
		Pos:      deferStmtPos,
		End:      0,
		Category: "line_breaks",
		Message:  "Line break before 'defer' statement is not allowed.",
	})
}
