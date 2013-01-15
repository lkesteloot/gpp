// Copyright 2013 Lawrence Kesteloot

package main

import (
	"go/ast"
	"go/token"
	"io/ioutil"
)

type template interface {
	Generate() ast.Stmt
}

type templateText struct {
	text string
}

func (t *templateText) Generate() ast.Stmt {
	return makeWriteStmt(t.text)
}

func loadTemplate(filename string) (template, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return &templateText{string(content)}, nil
}

func makeWriteStmt(text string) ast.Stmt {
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
						&ast.BasicLit{
							Kind: token.STRING,
							Value: "`" + text + "`",
						},
					},
				},
			},
		},
	}
	return e
}
