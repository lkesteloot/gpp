
package main

import (
	"io"
	"os"
)

// Dumps the index page of the website.
func index(f io.Writer) {
	name := "Lawrence"

	include("example/index.html")

	{
		f.Write(([]byte)(name))
	}
}

func main() {
	index(os.Stdout)
}
