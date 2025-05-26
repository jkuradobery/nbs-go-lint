package main

import (
	"golang.org/x/tools/go/analysis"

	"github.com/golangci/plugin-module-register/register"

	"github.com/jkuradobery/nbs-go-lint/line_breaks_analyzer"
	signature "github.com/jkuradobery/nbs-go-lint/multiline_signature_analyzer"
	"github.com/jkuradobery/nbs-go-lint/separator_analyzer"
)

type NbsAnalyzerPlugin struct{}

func (n NbsAnalyzerPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		line_breaks_analyzer.LineBreakAfterRbracket(),
		separator_analyzer.SeparatorAnalyzer(),
		signature.LineBreakAfterMultilineFunctionSignatureAnalyzer(),
	}, nil
}

func (n NbsAnalyzerPlugin) GetLoadMode() string {
	return register.LoadModeSyntax
}

func init() {
	register.Plugin(
		"nbs-go-lint",
		func(conf any) (register.LinterPlugin, error) {
			return &NbsAnalyzerPlugin{}, nil
		},
	)
}
