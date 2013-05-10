Gpp is a Go pre-processor that inlines template files and performs
minor transformations.

Installing
----------

Fetch with:

    go get github.com/lkesteloot/gpp

Run with:

    gpp input.gt >output.go
    go build output.go

Use the extension ".gt" for these files, though they're mostly Go
code. This avoids running into problems with commands like "go run \*.go".

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
        include(f, "example.html")
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

Raw Expressions
---------------

To avoid HTML escaping, add a slash at the front of the expression:

    Hello, {{/ highlightWord(name) }}!

Directives
----------

The template language supports three directives, all within {# #}. If statements:

    {# if expr #}
        Hello!
    {# end #}

For statements, which must be "range" statements:

    {# for _, name := range names #}
        Hello {{ name }}!
    {# end #}

With statements, defining a new variable only in this block:

    {# with name := person.getName() #}
        Hello {{ name }}!
    {# end #}

The "end" part can be followed by anything you want, which helps keep deeply
nested directives easier to read:

    {# for _, user := range users #}
        {# with name := user.getName() #}
            Hello {{ name }}!
        {# end with name #}
    {# end for user #}

Statements
----------

A plain statement is in {% %} and is usually a function call to include another
template:

    {% generateHeader(__out__, user) %}

The pseudo-variable \_\_out\_\_ will be replaced by whatever output
expression was passed in to this function (usually a bufio.Writer
pointer).

Static file reference
---------------------

Reference a static file using the {$ $} construct:

    <link rel="stylesheet" href="{$ /static/main.css $}">

This adds the hash of the file contents to the URL, guaranteeing that the
client will reload the file if its contents change. The hash is computed
at compile time:

    f.WriteString(`<link rel="stylesheet" href="`)
    f.WriteString(`/static/main.css?632349d89c`)
    f.WriteString(`">`)

The filename is rooted at the path specified by the -static command-line flag.

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
