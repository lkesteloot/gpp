// Copyright 2013 Lawrence Kesteloot

package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
)

func main() {
	// Get input file.
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Println("Must supply one input filename.")
		os.Exit(1)
	}
	inputFilename := flag.Arg(0)

	// We don't really need this.
	fset := token.NewFileSet()

	// Parse input file.
	f, err := parser.ParseFile(fset, inputFilename, nil, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Process the source file.
	p := NewPreprocessor()
	p.preprocess(f)

	// Print transformed source code.
	printer.Fprint(os.Stdout, fset, f)
}
