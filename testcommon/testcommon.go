package testcommon

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestdataDir(t *testing.T) string {
	t.Helper()
	_, testFilename, _, ok := runtime.Caller(1)
	require.True(t, ok)
	return filepath.Join(filepath.Dir(testFilename), "testdata")
}
