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

func parseFile(filename string) ast.Stmt {
	t, err := loadTemplate(filename)
	if err != nil {
		fmt.Printf("Cannot read file \"%s\" (%s)\n", filename, err)
		os.Exit(1)
	}

	return t.Generate()
}

func processNode(n ast.Node) bool {
	switch e := n.(type) {
	case *ast.BlockStmt:
		filename, index, ok := findIncludeCall(e)
		if ok {
			e.List[index] = parseFile(filename)
		}
	}

	return true
}

func main() {
	inputFilename := "example/input.go"

	fset := token.NewFileSet()

	// Parse input file.
	f, err := parser.ParseFile(fset, inputFilename, nil, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	/// ast.Print(fset, f)
	ast.Inspect(f, processNode)
	/// ast.Print(fset, f)

	// Print transformed source code.
	printer.Fprint(os.Stdout, fset, f)
}
