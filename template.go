// Copyright 2013 Lawrence Kesteloot

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
)

type template interface {
	Generate() ast.Stmt
}

type templateText struct {
	text string
}

func (t *templateText) Generate() ast.Stmt {
	return makeWriteStmt(&ast.BasicLit{Kind: token.STRING, Value: "`" + t.text + "`"})
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
	return makeWriteStmt(t.expr)
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
					fmt.Fprintf(os.Stderr, "Invalid expression \"%s\"\n", exprText)
					os.Exit(1)
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

func makeWriteStmt(expr ast.Expr) ast.Stmt {
	e := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "f",
				},
				Sel: &ast.Ident {
					Name: "Write",
				},
			},
			Args: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.ParenExpr{
						X: &ast.ArrayType{
							Elt: &ast.Ident {
								Name: "byte",
							},
						},
					},
					Args: []ast.Expr{
						expr,
					},
				},
			},
		},
	}
	return e
}
