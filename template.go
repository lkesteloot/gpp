// Copyright 2013 Lawrence Kesteloot

package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/lkesteloot/astutil"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
	"strings"
)

const (
	dumpExpression = false
)

// ---------------------------------------------------------------------------

type parserState int

const (
	// Plain text, outside directives.
	stateText parserState = iota
	stateTextOneOpenBrace

	// Expression, evaluated and inserted, HTML-escaped. Put a / right after
	// the {{ to not HTML-escape.
	stateExpression
	stateExpressionCloseBrace

	// Ifs and fors.
	stateStatement
	stateStatementPercent

	// Statement, usually a function call.
	stateDirective
	stateDirectiveHash

	// Reference to a static file. Adds the hash of the file as a query parameter
	// to bust the cache.
	stateStatic
	stateStaticDollar
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
	raw bool
}

func (t *templateExpr) Generate(outputExpr ast.Expr) ast.Stmt {
	expr := t.expr
	if !t.raw {
		expr = makeEscapeExpr(expr)
	}

	return makeWriteStmt(outputExpr, expr)
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

type templateIf struct {
	condition ast.Expr
	thenNode templateNode
	elseNode templateNode
}

func (t *templateIf) Generate(outputExpr ast.Expr) ast.Stmt {
	thenStmt := t.thenNode.Generate(outputExpr).(*ast.BlockStmt)
	var elseStmt ast.Stmt
	if t.elseNode != nil {
		elseStmt = t.elseNode.Generate(outputExpr)
	}

	return &ast.IfStmt{
		Cond: t.condition,
		Body: thenStmt,
		Else: elseStmt,
	}
}

type templateFor struct {
	key, value string
	rangeExpr ast.Expr
	body templateNode
}

func (t *templateFor) Generate(outputExpr ast.Expr) ast.Stmt {
	bodyStmt := t.body.Generate(outputExpr).(*ast.BlockStmt)

	r := &ast.RangeStmt{
		Key: &ast.Ident{
			Name: t.key,
		},
		Tok: token.DEFINE,
		X: t.rangeExpr,
		Body: bodyStmt,
	}

	// Value variable is optional.
	if t.value != "" {
		r.Value = &ast.Ident{
			Name: t.value,
		}
	}

	return r
}

// ---------------------------------------------------------------------------

func parseTemplate(contentString string) (templateNode, error) {
	var bStack []*templateBlock
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
			} else if ch == '#' {
				flushText()
				state = stateDirective
			} else if ch == '$' {
				flushText()
				state = stateStatic
			} else {
				segment = append(segment, '{')
				segment = append(segment, ch)
				state = stateText
			}
		case stateExpression:
			if ch == '}' {
				state = stateExpressionCloseBrace
			} else {
				segment = append(segment, ch)
			}
		case stateExpressionCloseBrace:
			if ch == '}' {
				exprText := string(segment)
				var raw = false
				if len(exprText) > 0 && exprText[0] == '/' {
					raw = true
					exprText = exprText[1:]
				}
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
				b.list = append(b.list, &templateExpr{expr, raw})
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
		case stateDirective:
			if ch == '#' {
				state = stateDirectiveHash
			} else {
				segment = append(segment, ch)
			}
		case stateDirectiveHash:
			if ch == '}' {
				directive := strings.TrimSpace(string(segment))
				segment = []rune{}

				// See what the directive is.
				if strings.HasPrefix(directive, "if ") {
					exprText := directive[3:]
					expr, err := parser.ParseExpr(exprText)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Invalid expression: %s\n", exprText)
						os.Exit(1)
					}
					bThen := &templateBlock{}
					b.list = append(b.list, &templateIf{
						condition: expr,
						thenNode: bThen,
						elseNode: nil,
					})
					bStack = append(bStack, b)
					b = bThen
				} else if strings.HasPrefix(directive, "for ") {
					stmtText := directive[4:]
					// Split at :=
					stmtFields := strings.SplitN(stmtText, ":=", 2)
					if len(stmtFields) != 2 {
						fmt.Fprintf(os.Stderr, "For loop must have assignment: %s\n", stmtText)
						os.Exit(1)
					}
					// Split LHS at comma.
					varFields := strings.Split(stmtFields[0], ",")
					if len(varFields) > 2 {
						fmt.Fprintf(os.Stderr,
							"LHS of for loop must have at most two variables: %s\n", stmtText)
						os.Exit(1)
					} else if len(varFields) == 1 {
						varFields = append(varFields, "")
					}
					// Get rid of "range" prefix.
					rhs := strings.TrimSpace(stmtFields[1])
					if !strings.HasPrefix(rhs, "range ") {
						fmt.Fprintf(os.Stderr, "For loop must be range: %s\n", rhs)
						os.Exit(1)
					}
					rhs = rhs[6:]
					rangeExpr, err := parser.ParseExpr(rhs)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Invalid expression: %s (%s)\n", rhs, err)
						os.Exit(1)
					}
					bForBody := &templateBlock{}
					b.list = append(b.list, &templateFor{
						key: varFields[0],
						value: varFields[1],
						rangeExpr: rangeExpr,
						body: bForBody,
					})
					bStack = append(bStack, b)
					b = bForBody
				} else if directive == "else" || strings.HasPrefix(directive, "else ") {
					if len(bStack) == 0 {
						fmt.Fprintf(os.Stderr, "Else without an if")
						os.Exit(1)
					}
					bLast := bStack[len(bStack)-1]
					if len(bLast.list) == 0 {
						fmt.Fprintf(os.Stderr, "Else without an if")
						os.Exit(1)
					}
					tIf, ok := bLast.list[len(bLast.list)-1].(*templateIf)
					if !ok {
						fmt.Fprintf(os.Stderr, "Else without an if")
						os.Exit(1)
					}
					bElse := &templateBlock{}
					tIf.elseNode = bElse
					b = bElse
				} else if directive == "end" || strings.HasPrefix(directive, "end ") {
					// End of "if" or "for".
					if len(bStack) == 0 {
						fmt.Fprintf(os.Stderr, "Too many ends")
						os.Exit(1)
					}
					b = bStack[len(bStack)-1]
					bStack = bStack[:len(bStack)-1]
				} else {
					fmt.Fprintf(os.Stderr, "Invalid directive: %s\n", directive)
					os.Exit(1)
				}

				state = stateText
			} else {
				segment = append(segment, '#')
				segment = append(segment, ch)
				state = stateDirective
			}
		case stateStatic:
			if ch == '$' {
				state = stateStaticDollar
			} else {
				segment = append(segment, ch)
			}
		case stateStaticDollar:
			if ch == '}' {
				static := strings.TrimSpace(string(segment))

				if *staticPath != "" {
					// Get file contents.
					localFilename := *staticPath + "/" + static
					file, err := os.Open(localFilename)

					if err == nil {
						defer file.Close()

						// Compute hash.
						h := sha1.New()
						io.Copy(h, file)

						// Convert to string.
						hashString := fmt.Sprintf("%x", h.Sum(nil))

						// Clip, no sense in keeping it all.
						hashString = hashString[:10]

						// Add to static.
						static += "?" + hashString
					} else {
						fmt.Fprintf(os.Stderr, "File \"%s\" not found", localFilename)

					}
				}

				segment = ([]rune)(static)
				flushText();
				state = stateText
			} else {
				segment = append(segment, '$')
				segment = append(segment, ch)
				state = stateStatic
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
		fmt.Fprintf(os.Stderr, "Unmatched brace (%d)\n", state)
		os.Exit(1)
	}

	if len(bStack) > 0 {
		fmt.Fprintf(os.Stderr, "Unbalanced directive")
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
