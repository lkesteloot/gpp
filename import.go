// Copyright 2013 Lawrence Kesteloot

package main

import (
	"github.com/lkesteloot/astutil"
	"go/ast"
	"go/token"
)

type importChecker struct {
	pkgName string
	found bool
}

func (ic *importChecker) ProcessNode(node ast.Node) {
	// Nothing.
}

func (ic *importChecker) ProcessIdent(ident **ast.Ident) {
	// Nothing.
}

func (ic *importChecker) ProcessExpr(expr *ast.Expr) {
	switch e := (*expr).(type) {
	case *ast.SelectorExpr:
		i, ok := e.X.(*ast.Ident)
		if ok {
			if i.Name == ic.pkgName {
				ic.found = true
			}
		}
	}
}

func (ic *importChecker) ProcessStmt(stmt *ast.Stmt) {
	// Nothing.
}

func (ic *importChecker) ProcessDecl(decl *ast.Decl) {
	// Nothing.
}

func addImport(f *ast.File, pkgName string) {
	ic := &importChecker{pkgName, false}
	astutil.VisitNode(ic, f)

	if ic.found {
		// XXX See if it's already imported.

		importSpec := &ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind: token.STRING,
				Value: "\"" + pkgName + "\"",
			},
		}
		f.Decls = append([]ast.Decl{
			&ast.GenDecl{
				Tok: token.IMPORT,
				Specs: []ast.Spec{
					importSpec,
				},
			},
		}, f.Decls...)
	}
}
