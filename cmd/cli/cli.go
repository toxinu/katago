package main

import (
	"fmt"

	"github.com/toxinu/katago/backends"
	"github.com/toxinu/katago/downloader"
)

func main() {
	d := downloader.NewDownloader("mangafox")

	fmt.Println(" :: Backends")
	for _, backend := range backends.Backends {
		if backend.Name() == d.Backend.Name() {
			fmt.Println("     -", backend.Name(), "(enabled)")
		} else {
			fmt.Println("     -", backend.Name())
		}
	}

	mangas, err := d.Backend.Search("one piece")
	if err != nil {
		panic(err)
	}

	manga := mangas[0]
	fmt.Println(" :: Manga   =>", manga)

	chapters, err := d.Backend.Chapters(manga)
	if err != nil {
		panic(err)
	}
	chapter := chapters[0]
	fmt.Println(" :: Chapter =>", chapter)

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
	// 	fmt.Println(page, URL)
	// }
}
