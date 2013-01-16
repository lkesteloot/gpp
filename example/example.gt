
// +build foo

package main

import (
	"bufio"
	"os"
)

func title(f *bufio.Writer, title string) {
	include(f, "example/title.html")
}

// Dumps the index page of the website.
func index(f *bufio.Writer) {
	name := "<b>Lawrence</b>"

	include(f, "example/index.html")
}

func main() {
	f := bufio.NewWriter(os.Stdout)
	index(f)
	f.Flush()
}
