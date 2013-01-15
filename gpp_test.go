// Copyright 2013 Lawrence Kesteloot

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"testing"
)

func stringToAst(input string) ast.Node {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", input, 0)
	if err != nil {
		panic(err)
	}

	return f
}

func astToString(node ast.Node) string {
	fset := token.NewFileSet()

	// Print to buffer.
	var b bytes.Buffer
	printer.Fprint(&b, fset, node)
	return b.String()
}

func cleanUp(input string) string {
	return astToString(stringToAst(input))
}

func compare(t *testing.T, input, expectedOutput string) {
	expectedOutput = cleanUp(expectedOutput)

	f := stringToAst(input)
	ast.Inspect(f, processNode)
	actualOutput := astToString(f)
	if actualOutput != expectedOutput {
		fmt.Println(actualOutput)
		fmt.Println(expectedOutput)
	}
}

func TestNoOp(t *testing.T) {
	input := `
package foo
func main() {
}
`
	compare(t, input, input)
}
