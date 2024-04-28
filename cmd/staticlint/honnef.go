package main

import (
	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/analysis/lint"
	"slices"
)

func lints2Analyzers(list []*lint.Analyzer) []*analysis.Analyzer {
	result := make([]*analysis.Analyzer, len(list))
	for i, v := range list {
		result[i] = v.Analyzer
	}
	return result
}

func lintPick(list []*lint.Analyzer, names ...string) []*analysis.Analyzer {
	maxLen := len(names)
	result := make([]*analysis.Analyzer, maxLen)
	i := 0
	for _, v := range list {
		if slices.Contains(names, v.Analyzer.Name) {
			result[i] = v.Analyzer
			if i++; i == maxLen {
				break
			}
		}
	}
	return result
}
