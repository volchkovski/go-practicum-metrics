// Package main contains the custom noexitmain analyzer for checking os.Exit usage in main functions.
package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// NoExitMainAnalyzer is a custom analyzer for checking os.Exit calls in main functions.
//
// The analyzer is designed to detect direct calls to os.Exit in the main function
// of main packages (package main). Such calls are considered undesirable because:
//
//  1. They make code less testable - impossible to test the main function
//  2. They complicate graceful shutdown - no way to properly close resources
//  3. They violate clean architecture principles - business logic mixed with system calls
//  4. They can lead to data loss - incomplete write operations
//
// The analyzer only works with packages named "main" and checks only functions named "main".
//
// # Examples of detected issues:
//
//	func main() {
//	    if err := run(); err != nil {
//	        os.Exit(1) // ⚠️ Will be detected by the analyzer
//	    }
//	}
//
// # Recommended alternative:
//
//	func main() {
//	    if err := run(); err != nil {
//	        log.Fatal(err) // ✅ Recommended approach
//	        // or return without os.Exit
//	    }
//	}
var NoExitMainAnalyzer = &analysis.Analyzer{
	Name: "noexitmain",
	Doc:  "checks that os.Exit is not called directly from main function in main package",
	Run:  run,
}

// run executes the main logic of the noexitmain analyzer.
//
// The function analyzes the AST tree of the package and looks for os.Exit calls in main functions.
//
// Algorithm:
//  1. Checks that the analyzed package has the name "main"
//  2. Iterates through all files in the package
//  3. For each file, examines all declarations
//  4. Finds functions named "main"
//  5. Recursively traverses the AST of the main function body
//  6. Looks for function calls of the form os.Exit(...)
//  7. Reports an error when such calls are found
//
// The analyzer recognizes only direct calls to os.Exit, but not aliases or
// indirect calls through function variables.
//
// Parameters:
//
//	pass - analyzer context containing information about the package, files, and AST
//
// Returns:
//
//	any - analysis result (always nil for this analyzer)
//	error - analysis execution error (nil on successful execution)
func run(pass *analysis.Pass) (any, error) {
	// Analyze only main packages (package main)
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	// Go through all files in the package
	for _, f := range pass.Files {
		// Check all declarations in the file
		for _, decl := range f.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			// Look for main functions only
			if funcDecl.Name.Name != "main" {
				continue
			}

			// Recursively traverse the main function body
			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				callExpr, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				// Check if the call is a selector (package.function)
				selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				// Check that it's an identifier (package name)
				ident, ok := selExpr.X.(*ast.Ident)
				if !ok {
					return true
				}

				// Check for os.Exit call
				if ident.Name == "os" && selExpr.Sel.Name == "Exit" {
					pass.Reportf(callExpr.Pos(), "direct call to os.Exit in main function is not allowed")
				}
				return true
			})
		}
	}
	return nil, nil
}
