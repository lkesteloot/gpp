Gpp is a Go pre-processor that inlines template files.

Fetch with:

    go get github.com/lkesteloot/gpp

Run with:

    gpp input.go >output.go
    go build output.go

Where the input file is a Go file with special directives:

    package main

    import (
        "io"
        "os"
    )

    // Dumps the index page of the website.
    func index(f io.Writer) {
        include("example.html")
    }

    func main() {
        index(os.Stdout)
    }

The fake function include() will be replaced by code that outputs the
specified text file to the Writer.
