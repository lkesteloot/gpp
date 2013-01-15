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

func parseIncludeCall(s ast.Stmt) (filename string, ok bool) {
	e, ok := s.(*ast.ExprStmt)
	if ok {
		c, ok := e.X.(*ast.CallExpr)
		if ok {
			i, ok := c.Fun.(*ast.Ident)
			if ok {
				if i.Name == "include" {
					arg0, ok := c.Args[0].(*ast.BasicLit)
					if ok && arg0.Kind == token.STRING {
						filename := arg0.Value
						// Strip out quotes.
						filename = filename[1:len(filename) - 1]
						return filename, true
					}
				}
			}
		}
	}
	ok = false
	return
}

func findIncludeCall(b *ast.BlockStmt) (filename string, index int, ok bool) {
	var s ast.Stmt

	for index, s = range b.List {
		filename, ok = parseIncludeCall(s)
		if ok {
			return
		}
	}
	ok = false
	return
}

func parseFile(filename string) *ast.BlockStmt {
	t, err := loadTemplate(filename)
	if err != nil {
		fmt.Printf("Cannot read file \"%s\" (%s)\n", filename, err)
		os.Exit(1)
	}

	b := &ast.BlockStmt{}

	b.List = append(b.List, t.Generate())

	return b
}

func processNode(n ast.Node) bool {
	b, ok := n.(*ast.BlockStmt)
	if ok {
		filename, index, ok := findIncludeCall(b)
		if ok {
			b.List[index] = parseFile(filename)
			return false
		}
	}

	return true
}

func main() {
	inputFilename := "example/input.go"

	fset := token.NewFileSet() // positions are relative to fset

	// Parse file.
	f, err := parser.ParseFile(fset, inputFilename, nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return
	}

	/// ast.Print(fset, f)
	ast.Inspect(f, processNode)
	/// ast.Print(fset, f)

	printer.Fprint(os.Stdout, fset, f)
}
