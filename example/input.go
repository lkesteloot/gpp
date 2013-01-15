
package main

import (
	"bufio"
	"os"
)

func title(f *bufio.Writer, title string) {
	include("example/title.html")
}

// Dumps the index page of the website.
func index(f *bufio.Writer) {
	name := "Lawrence"

	include("example/index.html")
}

func main() {
	f := bufio.NewWriter(os.Stdout)
	index(f)
	f.Flush()
}
