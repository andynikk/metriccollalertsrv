package banosexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var BanOsExit = &analysis.Analyzer{
	Name: "banosexit",
	Doc:  "Ban used os.Exit()",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	funcMain := false

	for _, f := range pass.Files {
		funcMain = false
		ast.Inspect(f, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl:
				funcMain = x.Name.Name == "main"
			case *ast.ExprStmt:
				banOsExit(x, pass, funcMain)
			}
			return true
		})
	}
	return nil, nil
}

func banOsExit(es *ast.ExprStmt, pass *analysis.Pass, funcMain bool) {
	if !funcMain {
		return
	}

	valCE, ok := es.X.(*ast.CallExpr)
	if !ok {
		return
	}
	valSE, ok := valCE.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	valI, ok := valSE.X.(*ast.Ident)
	if !ok {
		return
	}
	if valI.Name == "os" && valSE.Sel.Name == "Exit" {
		pass.Reportf(es.Pos(), "ban used os.Exit()")
	}
}
