package noexit

import (
	"golang.org/x/tools/go/analysis/analysistest"
	"testing"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "./...")
}
