package plugin

import (
	"github.com/jkuradobery/nbs-go-lint/line_breaks_analyzer"
	"golang.org/x/tools/go/analysis"

	signature "github.com/jkuradobery/nbs-go-lint/multiline_signature_analyzer"
)

func New(conf any) ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		signature.LineBreakAfterMultilineFunctionSignatureAnalyzer(),
		line_breaks_analyzer.LineBreakAfterRbracket(),
	}, nil
}
