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
        <p>Hello, `))
		f.Write(([]byte)(name))
		f.Write(([]byte)(`!</p>
    </body>
</html>
`))
	}

}

func main() {
	index(os.Stdout)
}
