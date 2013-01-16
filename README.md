Gpp is a Go pre-processor that inlines template files and performs
minor transformations.

Installing
----------

Fetch with:

    go get github.com/lkesteloot/gpp

Run with:

    gpp input.go >output.go
    go build output.go

Templates
---------

The input file is a Go file with special directives:

    package main

    import (
        "bufio"
        "os"
    )

    // Dumps the index page of the website.
    func index(f *bufio.Writer) {
        include("example.html")
    }

    func main() {
        f := bufio.NewWriter(os.Stdout)
        index(f)
        f.Flush()
    }

The fake function include() will be replaced by code that outputs the
specified text file to the Writer.

Expressions
-----------

A template can include a string expression within double curly braces:

    Hello, {{ name }}!

The expression must be a string. It will be HTML-escaped automatically. The
expression may reference local variables, method parameters, and generally
anything that is accessible from the include() call.

String formatting
-----------------

A new simple syntax is provided for converting non-strings to strings:

    "%03d: %02X"(a, b)

This is a function call where the function name is a string literal. This
is converted to this:

    fmt.Sprintf("%03d: %02X", a, b)

and the "fmt" module is automatically imported if necessary. This syntax
is useful in the {{ expression }} expansion of templates, since that only
takes strings, but it's available anywhere in your Go code.
