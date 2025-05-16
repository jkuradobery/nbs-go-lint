package line_breaks_analyzer

import (
	"testing"

	"github.com/jkuradobery/nbs-go-lint/testcommon"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestTodoAnalyzer(t *testing.T) {
	analysistest.TestData()
	analysistest.Run(
		t,
		testcommon.TestdataDir(t),
		LineBreakAfterRbracket(),
		"example/",
	)
}
