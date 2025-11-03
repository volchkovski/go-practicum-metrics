// Package main implements a static analyzer (multichecker) for Go code analysis.
//
// Staticlint is a comprehensive static analysis tool that combines various
// analyzers to check code quality and detect potential issues.
//
// # Included Analyzers
//
// ## Standard analyzers from golang.org/x/tools/go/analysis/passes:
//   - printf: checks correctness of format strings in printf-family functions
//   - shadow: detects variables that shadow variables from outer scope
//
// ## Analyzers from honnef.co/go/tools:
//   - staticcheck (SA*): static error checks and potential problem detection
//   - simple: code simplification suggestions
//   - stylecheck: Go coding style convention checks
//   - quickfix: quick fixes for detected issues
//
// ## Custom analyzers:
//   - noexitmain: checks that os.Exit is not called directly from main function
//
// # Execution Mechanism
//
// Multichecker works as follows:
//  1. Collects all analyzers into a single array
//  2. For staticcheck, filters only analyzers with "SA" prefix
//  3. Passes all analyzers to multichecker.Main for parallel execution
//  4. Each analyzer works independently on the AST tree of Go files
//  5. Results from all analyzers are combined and presented to the user
//
// # Usage
//
// Running the analyzer:
//
//	go run cmd/staticlint/main.go ./...
//	go run cmd/staticlint/main.go path/to/package
//
// Multichecker parameters can be passed via command line flags.
// Detailed help can be obtained using the -help flag.
package main

import (
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// main collects all analyzers and runs the multichecker.
//
// The function performs the following actions:
//  1. Creates an array to store all analyzers
//  2. Adds standard analyzers (printf, shadow)
//  3. Adds first analyzers from each honnef.co/go/tools category
//  4. Adds custom NoExitMainAnalyzer
//  5. Iterates through all staticcheck analyzers and adds only those starting with "SA"
//  6. Runs multichecker with the collected set of analyzers
//
// Staticcheck analyzers with "SA" prefix include checks for:
//   - SA1xxx: various diagnostics
//   - SA2xxx: concurrency issues
//   - SA3xxx: testing issues
//   - SA4xxx: code simplification
//   - SA5xxx: correctness issues
//   - SA6xxx: performance issues
//   - SA9xxx: dubious constructs
func main() {
	var analyzers []*analysis.Analyzer
	analyzers = append(
		analyzers,
		printf.Analyzer,                  // Check format strings in printf/scanf
		shadow.Analyzer,                  // Check variable shadowing
		simple.Analyzers[0].Analyzer,     // Code simplification suggestions
		stylecheck.Analyzers[0].Analyzer, // Coding style checks
		quickfix.Analyzers[0].Analyzer,   // Quick fixes
		NoExitMainAnalyzer,               // Custom check for os.Exit in main
	)

	// Add all staticcheck analyzers with "SA" prefix
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// Run multichecker with collected analyzers
	multichecker.Main(analyzers...)
}
