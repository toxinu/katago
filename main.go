package main

import (
	"github.com/toxinu/karago/backends"
)

func main() {
	d := backends.NewDownloader("mangafox")
	mangas := d.Search("one piece")
}
