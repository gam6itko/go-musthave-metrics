package noexit

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

// Analyzer не приемлет когда в функции main пакета main имеется вызов `os.Exit`.
var Analyzer = &analysis.Analyzer{
	Name: "noexit",
	Doc:  "there shouldn't be a exit from main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.String() != "main" {
			continue
		}

		checkExit := func(fn *ast.FuncDecl) {
			ast.Inspect(fn, func(node ast.Node) bool {
				if x, ok := node.(*ast.CallExpr); ok {
					if x2, ok2 := x.Fun.(*ast.SelectorExpr); ok2 {
						if x3, ok3 := x2.X.(*ast.Ident); ok3 {
							if x3.Name == "os" && x2.Sel.Name == "Exit" {
								pass.Reportf(node.Pos(), "no need to exit")
							}
						}
					}
				}
				return true
			})
		}

		//функцией ast.Inspect проходим по всем узлам AST
		ast.Inspect(file, func(node ast.Node) bool {
			if fn, ok := node.(*ast.FuncDecl); ok {
				if fn.Name.Name == "main" {
					checkExit(fn)
				}
			}

			return true
		})
	}
	return nil, nil
}
