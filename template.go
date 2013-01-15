// Copyright 2013 Lawrence Kesteloot

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
)

const (
	dumpExpression = false
)

type template interface {
	Generate() ast.Stmt
}

type templateText struct {
	text string
}

func (t *templateText) Generate() ast.Stmt {
	expr := &ast.BasicLit{Kind: token.STRING, Value: "`" + t.text + "`"}
	return makeWriteStmt(expr, false)
}

type templateBlock struct {
	list []template
}

func (t *templateBlock) Generate() ast.Stmt {
	b := &ast.BlockStmt{}

	for _, e := range t.list {
		b.List = append(b.List, e.Generate())
	}

	return b
}

type templateExpr struct {
	expr ast.Expr
}

func (t *templateExpr) Generate() ast.Stmt {
	return makeWriteStmt(t.expr, true)
}

func loadTemplate(filename string) (template, error) {
	contentBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	content := string(contentBytes)
	b := &templateBlock{}

	for {
		i := strings.Index(content, "{{")
		if i >= 0 {
			b.list = append(b.list, &templateText{content[:i]})

			content = content[i+2:]
			i = strings.Index(content, "}}")
			if i == -1 {
				fmt.Fprintln(os.Stderr, "Unmatched {{")
				os.Exit(1)
			} else {
				exprText := content[:i]
				expr, err := parser.ParseExpr(exprText)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Invalid expression: %s\n", exprText)
					os.Exit(1)
				}
				if dumpExpression {
					fset := token.NewFileSet()
					printer.Fprint(os.Stderr, fset, expr)
					fmt.Fprintln(os.Stderr)
				}
				b.list = append(b.list, &templateExpr{expr})
				content = content[i+2:]
			}
		} else {
			b.list = append(b.list, &templateText{content})
			break
		}
	}

	return b, nil
}

func makeWriteStmt(expr ast.Expr, escape bool) ast.Stmt {
	if escape {
		expr = &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "html",
				},
				Sel: &ast.Ident {
					Name: "EscapeString",
				},
			},
			Args: []ast.Expr{
				expr,
			},
		}
	}

	e := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "f",
				},
				Sel: &ast.Ident {
					Name: "WriteString",
				},
			},
			Args: []ast.Expr{
				expr,
			},
		},
	}
	return e
}
