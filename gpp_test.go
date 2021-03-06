// Copyright 2013 Lawrence Kesteloot

package main

import (
	"bytes"
	"errors"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"testing"
)

func stringToAst(input string) *ast.File {
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
	fakeFiles := map[string]string{
		"simple": "con{tent",
		"hello": "Hello {{ name }}!",
	}

	// Process input.
	f := stringToAst(input)
	p := NewPreprocessor()
	p.readFile = func(filename string) (string, error) {
		content, ok := fakeFiles[filename]
		if !ok {
			return "", errors.New("file not found")
		}
		return content, nil
	}
	p.preprocess(f)
	actualOutput := astToString(f)

	// Clean up output.
	expectedOutput = cleanUp(expectedOutput)

	if actualOutput != expectedOutput {
		t.Errorf("Different outputs (%s instead of expected %s)", actualOutput, expectedOutput)
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

func TestFormatOperator(t *testing.T) {
	input := `
		package foo
		func main() {
			x := "%3d: %02X"(5, 6)
		}
		`
	output := `
		package foo
		import "fmt"
		func main() {
			x := fmt.Sprintf("%3d: %02X", 5, 6)
		}
		`
	compare(t, input, output)
}

func TestSimpleInclude(t *testing.T) {
	input := `
		package foo
		func main() {
			include(f, "simple")
		}
		`
	output := `
		package foo
		func main() {
			{
				f.WriteString(` + "`con{tent`" + `)
			}
		}
		`
	compare(t, input, output)
}

func TestExpressionInclude(t *testing.T) {
	input := `
		package foo
		func main() {
			name := "Fred"
			include(f, "hello")
		}
		`
	output := `
		package foo
		import "html"
		func main() {
			name := "Fred"
			{
				f.WriteString(` + "`Hello `" + `)
				f.WriteString(html.EscapeString(name))
				f.WriteString(` + "`!`" + `)
			}
		}
		`
	compare(t, input, output)
}
