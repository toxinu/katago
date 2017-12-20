package main

import (
	"fmt"

	"github.com/toxinu/katago/backends"
)

func main() {
	d := backends.NewDownloader("mangafox")

	mangas, err := d.Backend.Search("one piece")
	if err != nil {
		panic(err)
	}

	manga := mangas[0]
	fmt.Println(" :: Manga   => #%v", manga)

	chapters, err := d.Backend.Chapters(manga)
	if err != nil {
		panic(err)
	}
	chapter := chapters[0]
	fmt.Println(" :: Chapter => #%v", chapter)

	err = d.Download(manga, chapter, "./mangas")
	if err != nil {
		panic(err)
	}
	// pages, err := d.Backend.Pages(chapter)
	// if err != nil {
	// 	panic(err)
	// }
	// for i := 0; i < len(pages); i++ {
	// 	page := pages[i]
	// 	URL, err := d.Backend.PageImageURL(page)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println("%#v %s", page, URL)
	// }
}
