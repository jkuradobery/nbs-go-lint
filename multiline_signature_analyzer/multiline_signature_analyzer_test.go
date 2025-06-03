package multiline_signature_analyzer

import (
	"github.com/jkuradobery/nbs-go-lint/testcommon"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestLineBreakAfterMultilineFunctionSignatureAnalyzer(t *testing.T) {
	analysistest.TestData()
	analysistest.Run(
		t,
		testcommon.TestdataDir(t),
		LineBreakAfterMultilineFunctionSignatureAnalyzer(),
		"example/",
	)
}
