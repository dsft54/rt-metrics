package main

import (
	"flag"
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

var mode = flag.String("mode", "all", "builtin - for passes analizers, static - for staticcheck, exits - for os.Exit in main, all by default")

func addBuiltinPasses(aP []*analysis.Analyzer) []*analysis.Analyzer {
	passes := []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		inspect.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		usesgenerics.Analyzer,
	}
	aP = append(aP, passes...)
	return aP
}

func addStaticCheck(aP []*analysis.Analyzer) []*analysis.Analyzer {
	for _, v := range staticcheck.Analyzers {
		aP = append(aP, v.Analyzer)
	}
	for _, v := range simple.Analyzers {
		if strings.Contains(v.Analyzer.Name, "S1000") {
			aP = append(aP, v.Analyzer)
		}
	}
	for _, v := range stylecheck.Analyzers {
		if strings.Contains(v.Analyzer.Name, "ST1000") {
			aP = append(aP, v.Analyzer)
		}
	}
	for _, v := range quickfix.Analyzers {
		if strings.Contains(v.Analyzer.Name, "QF1001") {
			aP = append(aP, v.Analyzer)
		}
	}
	return aP
}

var osExitAnalyzer = &analysis.Analyzer{
	Name: "osExit",
	Doc:  "check for os.Exit in main func",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, f := range pass.Files {
		packName := f.Name
		funcName := ""
		if packName.Name == "main" {
			ast.Inspect(f, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.CallExpr:
					res := fmt.Sprintf("%s", x.Fun)
					if res == "&{os Exit}" && funcName == "main" {
						pass.Reportf(x.Pos(), "os.Exit usage in main.go")
					}

				case *ast.FuncDecl:
					funcName = x.Name.Name
				}
				return true
			})
		}
	}
	return nil, nil
}

func addCustomCheck(aP []*analysis.Analyzer) []*analysis.Analyzer {
	aP = append(aP, osExitAnalyzer)
	return aP
}

func main() {
	flag.Parse()
	analyserPool := make([]*analysis.Analyzer, 0)
	switch *mode {
	case "all":
		analyserPool = addBuiltinPasses(analyserPool)
		analyserPool = addStaticCheck(analyserPool)
		analyserPool = addCustomCheck(analyserPool)
	case "builtin":
		analyserPool = addBuiltinPasses(analyserPool)
	case "static":
		analyserPool = addStaticCheck(analyserPool)
	case "exits":
		analyserPool = addCustomCheck(analyserPool)
	}
	multichecker.Main(
		analyserPool...,
	)
}
