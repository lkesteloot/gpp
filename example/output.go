package main

import (
	"bufio"
	"os"
)

func title(f *bufio.Writer, title string) {
	{
		f.WriteString(`<title>`)
		f.WriteString(title)
		f.WriteString(`</title>
`)
	}

}

// Dumps the index page of the website.
func index(f *bufio.Writer) {
	name := "Lawrence"
	{
		f.WriteString(`<!DOCTYPE html>
<html>
    <body>
        <p>Hello, `)
		f.WriteString(name)
		f.WriteString(`!</p>
    </body>
</html>
`)
	}

}

func main() {
	f := bufio.NewWriter(os.Stdout)
	index(f)
	f.Flush()
}
