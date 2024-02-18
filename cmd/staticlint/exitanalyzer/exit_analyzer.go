package exitanalyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var OSExitAnalyzer = &analysis.Analyzer{
	Name: "exitAnalyzer",
	Doc:  "Checks if there any os.Exit implementations in code",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			ast.Inspect(file, func(n ast.Node) bool {
				if isMainFunc(n) {
					switch x := n.(type) {
					case *ast.FuncDecl:
						ast.Inspect(x.Body, func(n ast.Node) bool {
							switch x := n.(type) {
							case *ast.CallExpr:
								if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
									if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "os" {
										if sel.Sel.Name == "Exit" {
											pass.Reportf(n.Pos(), "found os.Exit")
										}
									}
								}
							}
							return true
						})
					}
					return true
				}
				return true
			})
		}
	}
	return nil, nil
}

func isMainFunc(n ast.Node) bool {
	switch x := n.(type) {
	case *ast.FuncDecl:
		if x.Name.Name == "main" {
			return true
		}
	}
	return false
}
