package main

import (
	"io"
	"os"
)

// Dumps the index page of the website.
func index(f io.Writer) {
	name := "Lawrence"
	{
		f.Write(([]byte)(`<!DOCTYPE html>
<html>
    <body>
        <p>Hello, {{name}}!</p>
    </body>
</html>
`))
	}
	{
		f.Write(([]byte)(name))
	}
}

func main() {
	index(os.Stdout)
}
