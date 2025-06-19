package main

import (
	"go-metric-svc/internal/analyzer"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	saAnalyzers := staticcheck.Analyzers
	analyzers := []*analysis.Analyzer{
		analyzer.Analyzer,
	}

	for _, saAnalyzer := range saAnalyzers {
		analyzers = append(analyzers, saAnalyzer.Analyzer)
	}

	multichecker.Main(analyzers...)
}
