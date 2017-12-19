package main

import (
	"github.com/toxinu/katago/backends"
)

func main() {
	d := backends.NewDownloader("mangafox")
	mangas := d.Search("one piece")
}
