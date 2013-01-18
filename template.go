// Copyright 2013 Lawrence Kesteloot

package main

import (
	"fmt"
	"github.com/lkesteloot/astutil"
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
	stateTextOneOpenBrace
	stateExpression
	stateExpressionOneCloseBrace
	stateStatement
	stateStatementPercent
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

type templateStmt struct {
	stmt *ast.ExprStmt
}

type identifierSubstitutor struct {
	id string
	expr ast.Expr
}

// For astutil.Visitor:
func (subst *identifierSubstitutor) ProcessNode(node ast.Node) {}
func (subst *identifierSubstitutor) ProcessExpr(expr *ast.Expr) {
	e, ok := (*expr).(*ast.Ident)
	if ok && e.Name == subst.id {
		*expr = subst.expr
	}
}
func (subst *identifierSubstitutor) ProcessStmt(stmt *ast.Stmt) {}

func (t *templateStmt) Generate(outputExpr ast.Expr) ast.Stmt {
	// The statement can use the pseudo variable __out__ to mean the output
	// expression. Substitute it here, duplicating the original first.
	var stmt ast.Stmt = astutil.DuplicateExprStmt(t.stmt)

	// Do the substitution.
	subst := identifierSubstitutor{
		id: "__out__",
		expr: outputExpr,
	}
	astutil.VisitStmt(&subst, &stmt)

	return stmt
}

// ---------------------------------------------------------------------------

func parseTemplate(contentString string) (templateNode, error) {
	b := &templateBlock{}

	content := ([]rune)(contentString)
	state := stateText

	// Accumulated text so far in this state.
	segment := []rune{}

	flushText := func() {
		if len(segment) > 0 {
			b.list = append(b.list, &templateText{string(segment)})
			segment = []rune{}
		}
	}

	for _, ch := range content {
		switch state {
		case stateText:
			if ch == '{' {
				state = stateTextOneOpenBrace
			} else {
				segment = append(segment, ch)
			}
		case stateTextOneOpenBrace:
			if ch == '{' {
				flushText()
				state = stateExpression
			} else if ch == '%' {
				flushText()
				state = stateStatement
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
		case stateStatement:
			if ch == '%' {
				state = stateStatementPercent
			} else {
				segment = append(segment, ch)
			}
		case stateStatementPercent:
			if ch == '}' {
				stmtText := string(segment)
				segment = []rune{}
				stmt, err := parser.ParseExpr(stmtText)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Invalid statement: %s\n", stmtText)
					os.Exit(1)
				}
				if dumpExpression {
					fset := token.NewFileSet()
					printer.Fprint(os.Stderr, fset, stmt)
					fmt.Fprintln(os.Stderr)
				}
				b.list = append(b.list, &templateStmt{&ast.ExprStmt{X:stmt}})
				state = stateText
			} else {
				segment = append(segment, '%')
				segment = append(segment, ch)
				state = stateStatement
			}
		}
	}

	switch state {
	case stateText:
		flushText()
	case stateTextOneOpenBrace:
		segment = append(segment, '{')
		flushText()
	default:
		fmt.Fprintln(os.Stderr, "Unmatched brace")
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
