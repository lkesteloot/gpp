// Copyright 2013 Lawrence Kesteloot

package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"io/ioutil"
	"os"
)

type preprocessor struct {
	printTree bool
	readFile func(filename string) (string, error)
}

func NewPreprocessor() *preprocessor {
	p := &preprocessor{
		readFile: func(filename string) (string, error) {
			bytes, err := ioutil.ReadFile(filename)
			if err != nil {
				return "", err
			}
			return string(bytes), nil
		},
	}

	return p
}

func (p *preprocessor) parseIncludeCall(e *ast.ExprStmt) (outputExpr ast.Expr, filename string, ok bool) {
	c, ok := e.X.(*ast.CallExpr)
	if ok {
		i, ok := c.Fun.(*ast.Ident)
		if ok {
			if i.Name == "include" {
				if len(c.Args) != 2 {
					fmt.Fprintf(os.Stderr, "the include() function takes two arguments")
					os.Exit(1)
				}

				outputExpr = c.Args[0]

				filenameArg, ok := c.Args[1].(*ast.BasicLit)
				if ok && filenameArg.Kind == token.STRING {
					filename = filenameArg.Value
					// Strip out quotes.
					filename = filename[1:len(filename) - 1]
					return outputExpr, filename, true
				} else {
					fmt.Fprintf(os.Stderr, "the second parameter of include() must be a filename")
					os.Exit(1)
				}
			}
		}
	}
	ok = false
	return
}

func (p *preprocessor) parseFile(outputExpr ast.Expr, filename string) ast.Stmt {
	content, err := p.readFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read file \"%s\" (%s)\n", filename, err)
		os.Exit(1)
	}

	t, err := parseTemplate(content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot parse file \"%s\" (%s)\n", filename, err)
		os.Exit(1)
	}

	return t.Generate(outputExpr)
}

func (p *preprocessor) processNode(node ast.Node) {
	// Nothing.
}

func (p *preprocessor) processIdent(ident **ast.Ident) {
	// Nothing.
}

func (p *preprocessor) processExpr(expr *ast.Expr) {
	switch e := (*expr).(type) {
	case *ast.CallExpr:
		b, ok := e.Fun.(*ast.BasicLit)
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
				Args: append([]ast.Expr{e.Fun}, e.Args...),
			}
		}
	}
}

func (p *preprocessor) processStmt(stmt *ast.Stmt) {
	switch e := (*stmt).(type) {
	case *ast.ExprStmt:
		outputExpr, filename, ok := p.parseIncludeCall(e)
		if ok {
			*stmt = p.parseFile(outputExpr, filename)
		}
	}
}

func (p *preprocessor) processDecl(decl *ast.Decl) {
	// Nothing.
}

func (p *preprocessor) preprocess(f *ast.File) {
	fset := token.NewFileSet()

	if p.printTree {
		ast.Print(fset, f)
	}

	visitNode(p, f)
	if p.printTree {
		ast.Print(fset, f)
	}

	// Make sure these are imported. We may have added a reference to them.
	addImport(f, "fmt")
	addImport(f, "html")
}
