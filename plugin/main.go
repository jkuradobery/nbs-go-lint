package plugin

import (
	"golang.org/x/tools/go/analysis"

	signature "github.com/jkuradobery/nbs-go-lint/multiline_signature_analyzer"
)

func New(conf any) ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		signature.LineBreakAfterMultilineFunctionSignatureAnalyzer(),
	}, nil
}
