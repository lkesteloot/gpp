// Copyright 2013 Lawrence Kesteloot

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
)

const (
	dumpExpression = false
)

type template interface {
	Generate(outputExpr ast.Expr) ast.Stmt
}

type templateText struct {
	text string
}

func (t *templateText) Generate(outputExpr ast.Expr) ast.Stmt {
	expr := &ast.BasicLit{Kind: token.STRING, Value: "`" + t.text + "`"}
	return makeWriteStmt(outputExpr, expr)
}

type templateBlock struct {
	list []template
}

func (t *templateBlock) Generate(outputExpr ast.Expr) ast.Stmt {
	b := &ast.BlockStmt{}

	for _, e := range t.list {
		b.List = append(b.List, e.Generate(outputExpr))
	}

	return b
}

type templateExpr struct {
	expr ast.Expr
}

func (t *templateExpr) Generate(outputExpr ast.Expr) ast.Stmt {
	return makeWriteStmt(outputExpr, makeEscapeExpr(t.expr))
}

func parseTemplate(content string) (template, error) {
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

func makeEscapeExpr(expr ast.Expr) ast.Expr {
	return &ast.CallExpr{
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

func makeWriteStmt(outputExpr ast.Expr, expr ast.Expr) ast.Stmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: outputExpr,
				Sel: &ast.Ident {
					Name: "WriteString",
				},
			},
			Args: []ast.Expr{
				expr,
			},
		},
	}
}
