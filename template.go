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
	dumpExpression = false
)

// ---------------------------------------------------------------------------

type parserState int

const (
	stateText parserState = iota
	stateOneOpenBrace
	stateExpression
	stateExpressionOneCloseBrace
)

// ---------------------------------------------------------------------------

type templateNode interface {
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
	list []templateNode
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

// ---------------------------------------------------------------------------

func parseTemplate(contentString string) (templateNode, error) {
	b := &templateBlock{}

	content := ([]rune)(contentString)
	state := stateText

	// Accumulated text so far in this state.
	segment := []rune{}

	for _, ch := range content {
		switch state {
		case stateText:
			if ch == '{' {
				state = stateOneOpenBrace
			} else {
				segment = append(segment, ch)
			}
		case stateOneOpenBrace:
			if ch == '{' {
				if len(segment) > 0 {
					b.list = append(b.list, &templateText{string(segment)})
					segment = []rune{}
				}
				state = stateExpression
			} else {
				segment = append(segment, '{')
				segment = append(segment, ch)
				state = stateText
			}
		case stateExpression:
			if ch == '}' {
				state = stateExpressionOneCloseBrace
			} else {
				segment = append(segment, ch)
			}
		case stateExpressionOneCloseBrace:
			if ch == '}' {
				exprText := string(segment)
				segment = []rune{}
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
				state = stateText
			} else {
				segment = append(segment, '}')
				segment = append(segment, ch)
				state = stateExpression
			}
		}
	}

	if state == stateText {
		if len(segment) > 0 {
			b.list = append(b.list, &templateText{string(segment)})
		}
	} else {
		fmt.Fprintln(os.Stderr, "Unmatched {{")
		os.Exit(1)
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
