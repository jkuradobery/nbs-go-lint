package multiline_signature_analyzer

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

////////////////////////////////////////////////////////////////////////////////

func LineBreakAfterMultilineFunctionSignatureAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "LineBreakAfterMultilineFunctionSignatureAnalyzer",
		Doc:  "Checks for line breaks after multiline function signatures.",
		Run: func(pass *analysis.Pass) (any, error) {
			for _, file := range pass.Files {
				ast.Inspect(file, func(node ast.Node) bool {
					if function, ok := node.(*ast.FuncDecl); ok {
						processSingleFunction(pass, function, file)
					}

					return true
				})
			}

			return nil, nil
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

func findCommentBetweenFirstStatementAndLBrace(
	fset *token.FileSet,
	firstStmtPosition token.Position,
	lbracePosition token.Position,
	file *ast.File,
) (token.Position, bool) {

	for _, comment := range file.Comments {
		commentPosition := fset.Position(comment.Pos())
		// Skip comments that are not between the lbrace and the first statement
		if commentPosition.Line <= lbracePosition.Line {
			continue
		}

		if commentPosition.Line >= firstStmtPosition.Line {
			continue
		}

		return commentPosition, true
	}

	return token.Position{}, false
}

func processSingleFunction(
	pass *analysis.Pass,
	function *ast.FuncDecl,
	file *ast.File,
) {
	isMultiline := false
	params := function.Type.Params
	fset := pass.Fset
	opening := fset.Position(params.Opening)
	closing := fset.Position(params.Closing)
	if opening.Line < closing.Line {
		isMultiline = true
	}

	body := function.Body
	lbracePosition := fset.Position(body.Lbrace)
	if !isMultiline {
		return
	}

	stmt := body.Rbrace
	if len(body.List) > 0 {
		stmt = body.List[0].Pos()
	}

	firstStmtPosition := fset.Position(stmt)
	commentPosition, ok := findCommentBetweenFirstStatementAndLBrace(
		fset,
		firstStmtPosition,
		lbracePosition,
		file,
	)
	if ok {
		firstStmtPosition = commentPosition
	}

	difference := firstStmtPosition.Line - lbracePosition.Line
	if difference == 2 {
		return
	}

	message := "Line break after multiline " +
		"function signature is required"
	if difference > 2 {
		message = fmt.Sprintf(
			"Too many line breaks after the multiline "+
				"function signature: %d",
			difference-1,
		)
	}

	pass.Report(analysis.Diagnostic{
		Pos:      body.Lbrace,
		End:      stmt,
		Category: "line_breaks",
		Message:  message,
	})
}
