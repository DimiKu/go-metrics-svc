package analyzer

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "mycheck",
	Doc:  "Проверяет наличие os.Exit",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	foundFunc := "os.Exit"
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			callExpr, ok := node.(*ast.CallExpr)
			if ok {
				fun, ok := callExpr.Fun.(*ast.SelectorExpr)
				if ok {
					pkgName := fun.X.(*ast.Ident).Name
					funcName := fun.Sel.Name
					fullFuncName := pkgName + "." + funcName

					if foundFunc == fullFuncName {
						pass.Reportf(callExpr.Pos(), "usage of forbidden function: %s", fullFuncName)
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
