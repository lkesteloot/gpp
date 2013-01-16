// Copyright 2013 Lawrence Kesteloot

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	// Get input file.
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Println("Must supply input filenames.")
		flag.Usage()
		os.Exit(1)
	}

	for _, inputFilename := range flag.Args() {
		if !strings.HasSuffix(inputFilename, ".gt") {
			fmt.Println("Input files must have .gt extension.")
			flag.Usage()
			os.Exit(1)
		}
		outputFilename := inputFilename[:len(inputFilename)-3] + ".go"
		// XXX If output file exists, verify that it was created by us.

		// We don't really need this.
		fset := token.NewFileSet()

		// Parse input file.
		f, err := parser.ParseFile(fset, inputFilename, nil, 0)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Process the source file.
		p := NewPreprocessor()
		p.preprocess(f)

		// Write transformed source code.
		// XXX Don't write file if it's the same as what's on disk.
		var output bytes.Buffer
		printer.Fprint(&output, fset, f)
		ioutil.WriteFile(outputFilename, output.Bytes(), 0666)
	}
}
