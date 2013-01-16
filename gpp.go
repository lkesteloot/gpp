// Copyright 2013 Lawrence Kesteloot

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
)

const (
	printTree = false
)

func parseIncludeCall(s ast.Stmt) (filename string, ok bool) {
	e, ok := s.(*ast.ExprStmt)
	if ok {
		c, ok := e.X.(*ast.CallExpr)
		if ok {
			i, ok := c.Fun.(*ast.Ident)
			if ok {
				if i.Name == "include" {
					arg0, ok := c.Args[0].(*ast.BasicLit)
					if ok && arg0.Kind == token.STRING {
						filename := arg0.Value
						// Strip out quotes.
						filename = filename[1:len(filename) - 1]
						return filename, true
					}
				}
			}
		}
	}
	ok = false
	return
}

func parseFile(filename string) ast.Stmt {
	t, err := loadTemplate(filename)
	if err != nil {
		fmt.Printf("Cannot read file \"%s\" (%s)\n", filename, err)
		os.Exit(1)
	}

	return t.Generate()
}

type gppInspector bool

func (g gppInspector) processNode(node ast.Node) {
	// Nothing.
}

func (g gppInspector) processIdent(ident **ast.Ident) {
	// Nothing.
}

func (g gppInspector) processExpr(expr *ast.Expr) {
	switch e := (*expr).(type) {
	case *ast.BinaryExpr:
		if e.Op == token.REM {
			b, ok := e.X.(*ast.BasicLit)
			if ok && b.Kind == token.STRING {
				*expr = &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.Ident{
							Name: "fmt",
						},
						Sel: &ast.Ident {
							Name: "Sprintf",
						},
					},
					Args: []ast.Expr{
						e.X,
						e.Y,
					},
				}
			}
		}
	}
}

func (g gppInspector) processStmt(stmt *ast.Stmt) {
	switch e := (*stmt).(type) {
	case *ast.ExprStmt:
		filename, ok := parseIncludeCall(e)
		if ok {
			*stmt = parseFile(filename)
		}
	}
}

func (g gppInspector) processDecl(decl *ast.Decl) {
	// Nothing.
}

func proprocess(f *ast.File) {
	fset := token.NewFileSet()

	if printTree {
		ast.Print(fset, f)
	}
	var g gppInspector
	visitNode(g, f)
	if printTree {
		ast.Print(fset, f)
	}

	addImport(f, "fmt")
	addImport(f, "html")
}

func main() {
	inputFilename := "example/input.go"

	fset := token.NewFileSet()

	// Parse input file.
	f, err := parser.ParseFile(fset, inputFilename, nil, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	proprocess(f)

	// Print transformed source code.
	printer.Fprint(os.Stdout, fset, f)
}
