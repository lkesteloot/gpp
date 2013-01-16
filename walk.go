// Modification of the original walk.go file from the standard library,
// modified to pass around pointers to nodes.

package main

import (
	"fmt"
	"go/ast"
)

type visitor interface {
	processNode(node ast.Node)
	processIdent(ident **ast.Ident)
	processExpr(expr *ast.Expr)
	processStmt(stmt *ast.Stmt)
	processDecl(decl *ast.Decl)
}

func walkIdentList(v visitor, list []*ast.Ident) {
	for i, _ := range list {
		visitIdent(v, &list[i])
	}
}

func walkExprList(v visitor, list []ast.Expr) {
	for i, _ := range list {
		visitExpr(v, &list[i])
	}
}

func walkStmtList(v visitor, list []ast.Stmt) {
	for i, _ := range list {
		visitStmt(v, &list[i])
	}
}

func walkDeclList(v visitor, list []ast.Decl) {
	for i, _ := range list {
		visitDecl(v, &list[i])
	}
}

func visitBlockStmt(v visitor, stmt *ast.BlockStmt) {
	v.processNode(stmt)

	walkStmtList(v, stmt.List)
}

func visitArrayType(v visitor, a *ast.ArrayType) {
	v.processNode(a)

	if a.Len != nil {
		visitExpr(v, &a.Len)
	}
	visitExpr(v, &a.Elt)
}

func visitIdent(v visitor, ident **ast.Ident) {
	v.processIdent(ident)
}

func visitExpr(v visitor, expr *ast.Expr) {
	v.processExpr(expr)

	switch n := (*expr).(type) {
	case *ast.BadExpr, *ast.Ident, *ast.BasicLit:
		// nothing to do

	case *ast.Ellipsis:
		if n.Elt != nil {
			visitExpr(v, &n.Elt)
		}

	case *ast.FuncLit:
		visitNode(v, n.Type)
		visitBlockStmt(v, n.Body)

	case *ast.CompositeLit:
		if n.Type != nil {
			visitExpr(v, &n.Type)
		}
		walkExprList(v, n.Elts)

	case *ast.ParenExpr:
		visitExpr(v, &n.X)

	case *ast.SelectorExpr:
		visitExpr(v, &n.X)
		visitIdent(v, &n.Sel)

	case *ast.IndexExpr:
		visitExpr(v, &n.X)
		visitExpr(v, &n.Index)

	case *ast.SliceExpr:
		visitExpr(v, &n.X)
		if n.Low != nil {
			visitExpr(v, &n.Low)
		}
		if n.High != nil {
			visitExpr(v, &n.High)
		}

	case *ast.TypeAssertExpr:
		visitExpr(v, &n.X)
		if n.Type != nil {
			visitExpr(v, &n.Type)
		}

	case *ast.CallExpr:
		visitExpr(v, &n.Fun)
		walkExprList(v, n.Args)

	case *ast.StarExpr:
		visitExpr(v, &n.X)

	case *ast.UnaryExpr:
		visitExpr(v, &n.X)

	case *ast.BinaryExpr:
		visitExpr(v, &n.X)
		visitExpr(v, &n.Y)

	case *ast.KeyValueExpr:
		visitExpr(v, &n.Key)
		visitExpr(v, &n.Value)

	// Types
	case *ast.ArrayType:
		visitArrayType(v, n)

	default:
		fmt.Printf("ast.visitExpr: unexpected node type %T", n)
		panic("ast.visitExpr")
	}
}

func visitStmt(v visitor, stmt *ast.Stmt) {
	v.processStmt(stmt)

	switch n := (*stmt).(type) {
	case *ast.BadStmt:
		// nothing to do

	case *ast.DeclStmt:
		visitDecl(v, &n.Decl)

	case *ast.EmptyStmt:
		// nothing to do

	case *ast.LabeledStmt:
		visitIdent(v, &n.Label)
		visitStmt(v, &n.Stmt)

	case *ast.ExprStmt:
		visitExpr(v, &n.X)

	case *ast.SendStmt:
		visitExpr(v, &n.Chan)
		visitExpr(v, &n.Value)

	case *ast.IncDecStmt:
		visitExpr(v, &n.X)

	case *ast.AssignStmt:
		walkExprList(v, n.Lhs)
		walkExprList(v, n.Rhs)

	case *ast.GoStmt:
		visitNode(v, n.Call)

	case *ast.DeferStmt:
		visitNode(v, n.Call)

	case *ast.ReturnStmt:
		walkExprList(v, n.Results)

	case *ast.BranchStmt:
		if n.Label != nil {
			visitNode(v, n.Label)
		}

	case *ast.BlockStmt:
		visitBlockStmt(v, n)

	case *ast.IfStmt:
		if n.Init != nil {
			visitStmt(v, &n.Init)
		}
		visitExpr(v, &n.Cond)
		visitBlockStmt(v, n.Body)
		if n.Else != nil {
			visitStmt(v, &n.Else)
		}

	case *ast.CaseClause:
		walkExprList(v, n.List)
		walkStmtList(v, n.Body)

	case *ast.SwitchStmt:
		if n.Init != nil {
			visitStmt(v, &n.Init)
		}
		if n.Tag != nil {
			visitExpr(v, &n.Tag)
		}
		visitBlockStmt(v, n.Body)

	case *ast.TypeSwitchStmt:
		if n.Init != nil {
			visitStmt(v, &n.Init)
		}
		visitStmt(v, &n.Assign)
		visitBlockStmt(v, n.Body)

	case *ast.CommClause:
		if n.Comm != nil {
			visitStmt(v, &n.Comm)
		}
		walkStmtList(v, n.Body)

	case *ast.SelectStmt:
		visitBlockStmt(v, n.Body)

	case *ast.ForStmt:
		if n.Init != nil {
			visitStmt(v, &n.Init)
		}
		if n.Cond != nil {
			visitExpr(v, &n.Cond)
		}
		if n.Post != nil {
			visitStmt(v, &n.Post)
		}
		visitBlockStmt(v, n.Body)

	case *ast.RangeStmt:
		visitExpr(v, &n.Key)
		if n.Value != nil {
			visitExpr(v, &n.Value)
		}
		visitExpr(v, &n.X)
		visitBlockStmt(v, n.Body)

	default:
		fmt.Printf("ast.visitStmt: unexpected node type %T", n)
		panic("ast.visitStmt")
	}
}

func visitDecl(v visitor, decl *ast.Decl) {
	v.processDecl(decl)

	switch n := (*decl).(type) {
	case *ast.BadDecl:
		// nothing to do

	case *ast.GenDecl:
		if n.Doc != nil {
			visitNode(v, n.Doc)
		}
		for _, s := range n.Specs {
			visitNode(v, s)
		}

	case *ast.FuncDecl:
		if n.Doc != nil {
			visitNode(v, n.Doc)
		}
		if n.Recv != nil {
			visitNode(v, n.Recv)
		}
		visitIdent(v, &n.Name)
		visitNode(v, n.Type)
		if n.Body != nil {
			visitBlockStmt(v, n.Body)
		}

	default:
		fmt.Printf("ast.visitDecl: unexpected node type %T", n)
		panic("ast.visitDecl")
	}
}

func visitNode(v visitor, node ast.Node) {
	v.processNode(node)

	switch n := node.(type) {
	case *ast.BadExpr, *ast.Ident, *ast.BasicLit:
		// nothing to do

	// Comments and fields
	case *ast.Comment:
		// nothing to do

	case *ast.CommentGroup:
		for _, c := range n.List {
			visitNode(v, c)
		}

	case *ast.Field:
		if n.Doc != nil {
			visitNode(v, n.Doc)
		}
		walkIdentList(v, n.Names)
		visitExpr(v, &n.Type)
		if n.Tag != nil {
			visitNode(v, n.Tag)
		}
		if n.Comment != nil {
			visitNode(v, n.Comment)
		}

	case *ast.FieldList:
		for _, f := range n.List {
			visitNode(v, f)
		}

	// Types
	case *ast.ArrayType:
		visitArrayType(v, n)

	case *ast.StructType:
		visitNode(v, n.Fields)

	case *ast.FuncType:
		visitNode(v, n.Params)
		if n.Results != nil {
			visitNode(v, n.Results)
		}

	case *ast.InterfaceType:
		visitNode(v, n.Methods)

	case *ast.MapType:
		visitNode(v, n.Key)
		visitNode(v, n.Value)

	case *ast.ChanType:
		visitExpr(v, &n.Value)

	case *ast.ImportSpec:
		if n.Doc != nil {
			visitNode(v, n.Doc)
		}
		if n.Name != nil {
			visitIdent(v, &n.Name)
		}
		visitNode(v, n.Path)
		if n.Comment != nil {
			visitNode(v, n.Comment)
		}

	case *ast.ValueSpec:
		if n.Doc != nil {
			visitNode(v, n.Doc)
		}
		walkIdentList(v, n.Names)
		if n.Type != nil {
			visitExpr(v, &n.Type)
		}
		walkExprList(v, n.Values)
		if n.Comment != nil {
			visitNode(v, n.Comment)
		}

	case *ast.TypeSpec:
		if n.Doc != nil {
			visitNode(v, n.Doc)
		}
		visitIdent(v, &n.Name)
		visitExpr(v, &n.Type)
		if n.Comment != nil {
			visitNode(v, n.Comment)
		}


	// Files and packages
	case *ast.File:
		if n.Doc != nil {
			visitNode(v, n.Doc)
		}
		visitIdent(v, &n.Name)
		walkDeclList(v, n.Decls)
		for _, g := range n.Comments {
			visitNode(v, g)
		}
		// don't walk n.Comments - they have been
		// visited already through the individual
		// nodes

	case *ast.Package:
		for _, f := range n.Files {
			visitNode(v, f)
		}

	default:
		fmt.Printf("ast.visitNode: unexpected node type %T", n)
		panic("ast.visitNode")
	}
}
