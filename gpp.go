// Copyright 2013 Lawrence Kesteloot

package main

import (
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
)

func main() {
	inputFilename := "example/input.go"

	fset := token.NewFileSet()

	// Parse input file.
	f, err := parser.ParseFile(fset, inputFilename, nil, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	p := NewPreprocessor()
	p.preprocess(f)

	// Print transformed source code.
	printer.Fprint(os.Stdout, fset, f)
}
