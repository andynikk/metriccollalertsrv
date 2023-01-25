// Package staticlint: Статический анализ кода
//
// multichecker состоит из:
//
// стандартного статических анализаторов пакета golang.org/x/tools/go/analysis/passes;
//
// всех анализаторов класса SA пакета staticcheck.io;
//
// не менее одного анализатора остальных классов пакета staticcheck.io;
//
// публичные анализаторы commentmap, tagalyzer
//
// Создание бинарного файла анализатора "%patchapp%\go build ./cmd/staticlint/."
//
// Запускается из коммандной строки "%patchapp%\staticlint.exe ./..."
//
// Список анализаторов можно получить коммандой %patchapp%\staticlint.exe -help
//
// golang.org/x/tools/go/analysis
//
// asmdecl: Package asmdecl defines an Analyzer that reports mismatches between assembly files and Go declarations.
//
// assign: Package assign defines an Analyzer that detects useless assignments.
//
// atomic: Package atomic defines an Analyzer that checks for common mistakes using the sync/atomic package.
//
// atomicalign: Package atomicalign defines an Analyzer that checks for non-64-bit-aligned arguments to sync/atomic functions.
//
// bools: Package bools defines an Analyzer that detects common mistakes involving boolean operators.
//
// buildssa: Package buildssa defines an Analyzer that constructs the SSA representation of an error-free package and returns the set of all functions within it.
//
// buildtag: Package buildtag defines an Analyzer that checks build tags.
//
// cgocall: Package cgocall defines an Analyzer that detects some violations of the cgo pointer passing rules.
//
// composite: Package composite defines an Analyzer that checks for unkeyed composite literals.
//
// copylock: Package copylock defines an Analyzer that checks for locks erroneously passed by value.
//
// ctrlflow: Package ctrlflow is an analysis that provides a syntactic control-flow graph (CFG) for the body of a function.
//
// deepequalerrors: Package deepequalerrors defines an Analyzer that checks for the use of reflect.DeepEqual with error values.
//
// errorsas: The errorsas package defines an Analyzer that checks that the second arugment to errors.As is a pointer to a type implementing error.
//
// findcall: The findcall package defines an Analyzer that serves as a trivial example and test of the Analysis API.
//
//   - cmd/findcall: The findcall command runs the findcall analyzer.
//
// httpresponse: Package httpresponse defines an Analyzer that checks for mistakes using HTTP responses.
//
// inspect: Package inspect defines an Analyzer that provides an AST inspector (github.com/godaner/GCatch/tools/go/ast/inspect.Inspect) for the syntax trees of a package.
//
// loopclosure: Package loopclosure defines an Analyzer that checks for references to enclosing loop variables from within nested functions.
//
// lostcancel: Package lostcancel defines an Analyzer that checks for failure to call a context cancelation function.
//
//   - cmd/lostcancel: The lostcancel command applies the github.com/godaner/GCatch/tools/go/analysis/passes/lostcancel analysis to the specified packages of Go source code.
//
// nilfunc: Package nilfunc defines an Analyzer that checks for useless comparisons against nil.
//
// nilness: Package nilness inspects the control-flow graph of an SSA function and reports errors such as nil pointer dereferences and degenerate nil pointer comparisons.
//
//   - cmd/nilness: The nilness command applies the github.com/godaner/GCatch/tools/go/analysis/passes/nilness analysis to the specified packages of Go source code.
//
// pkgfact: The pkgfact package is a demonstration and test of the package fact mechanism.
//
// printf: check consistency of Printf format strings and arguments
//
// shadow:
//
//   - cmd/shadow: The shadow command runs the shadow analyzer.
//
// shift: Package shift defines an Analyzer that checks for shifts that exceed the width of an integer.
//
// stdmethods: Package stdmethods defines an Analyzer that checks for misspellings in the signatures of methods similar to well-known interfaces.
//
// structtag: Package structtag defines an Analyzer that checks struct field tags are well formed.
//
// tests: Package tests defines an Analyzer that checks for common mistaken usages of tests and examples.
//
// unmarshal: The unmarshal package defines an Analyzer that checks for passing non-pointer or non-interface types to unmarshal and decode functions.
//
//   - cmd/unmarshal: The unmarshal command runs the unmarshal analyzer.
//
// unreachable: Package unreachable defines an Analyzer that checks for unreachable code.
//
// unsafeptr: Package unsafeptr defines an Analyzer that checks for invalid conversions of uintptr to unsafe.Pointer.
//
// unusedresult: Package unusedresult defines an analyzer that checks for unused results of calls to certain pure functions.
//
// internal:
//
//   - analysisutil: Package analysisutil defines various helper functions used by two or more packages beneath go/analysis.
//
// ------------------------
//
// honnef.co/go/tools
//
// quickfix: Package quickfix contains analyzes that implement code refactorings.
//
// simple: Package simple contains analyzes that simplify code.
//
// staticcheck: Package staticcheck contains analyzes that find bugs and performance issues.
//
// staticcheck: Package staticcheck contains analyzes that find bugs and performance issues.
//
// ------------------------
//
// "github.com/gostaticanalysis/comment/passes/commentmap"
//
// commentmap: CommentMap utilities for static analysis in Go
//
// ------------------------
//
// "github.com/salihzain/tagalyzer"
//
// tagalyzer: Static analyzer to find missing tags in your Golang structs. This analyzer follows the Golang standardized
// way of creating static analyzers based on the Golang tools analysis package and is compatible
// with `go vet` command
//
// ------------------------
//
// "github.com/andynikk/advancedmetrics/internal/banosexit"
//
// BanOsExit - Ban used os.Exit()
package main

import (
	"strings"

	"github.com/andynikk/metriccollalertsrv/internal/banosexit"

	"github.com/gostaticanalysis/comment/passes/commentmap"
	"github.com/salihzain/tagalyzer"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	analyzers := []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		deepequalerrors.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		stdmethods.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		stringintconv.Analyzer,
		ifaceassert.Analyzer,
		banosexit.BanOsExit,
		commentmap.Analyzer,
		tagalyzer.Analyzer,
	}

	honnefTools := map[string]bool{
		"QF1003": true,
		"S1025":  true,
		"ST1017": true,
	}

	for _, val := range staticcheck.Analyzers {
		if strings.HasPrefix(val.Analyzer.Name, "SA") {
			analyzers = append(analyzers, val.Analyzer)
		}
	}

	for _, val := range stylecheck.Analyzers {
		if honnefTools[val.Analyzer.Name] {
			analyzers = append(analyzers, val.Analyzer)
		}
	}

	for _, val := range simple.Analyzers {
		if honnefTools[val.Analyzer.Name] {
			analyzers = append(analyzers, val.Analyzer)
		}
	}

	for _, val := range quickfix.Analyzers {
		if honnefTools[val.Analyzer.Name] {
			analyzers = append(analyzers, val.Analyzer)
		}
	}

	multichecker.Main(analyzers...)
}
