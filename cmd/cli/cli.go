package main

import (
	"fmt"
	"strconv"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/toxinu/katago/backends"
	"github.com/toxinu/katago/downloader"
)

var (
	backend = "mangafox"
	d, err  = downloader.NewDownloader(backend)
	manga   *backends.Manga
	p       *prompt.Prompt
	results []*backends.Manga
)

func executor(in string) {
	var err error

	splitted := strings.Fields(in)
	action := splitted[0]
	parameters := splitted[1:len(splitted)]

	switch action {
	case "backends":
		for _, backend := range backends.Backends {
			if backend.Name() == d.Backend.Name() {
				fmt.Println(backend.Name(), "(enabled)")
			} else {
				fmt.Println(backend.Name())
			}
		}
	case "use":
		if len(parameters) == 0 {
			fmt.Println("Error: backend name needed")
			return
		}
		backend = parameters[0]
		d, err = downloader.NewDownloader(backend)
		if err != nil {
			fmt.Println(err)
		}
	case "search":
		results, err = d.Backend.Search(strings.Join(parameters, " "))
		if err != nil {
			panic(err)
		}
		for index, result := range results {
			fmt.Printf("%d. %s\n", index, result.Name)
		}
	case "manga":
		var index int

		if len(parameters) == 0 {
			fmt.Println("Error: manga index needed")
			return
		}

		index, err = strconv.Atoi(parameters[0])
		if err != nil {
			fmt.Println("Error: invalid index (must be integer)")
			return
		}

		if index > len(results) {
			fmt.Println("Error: index out of range")
			return
		}

		manga = results[index]
		fmt.Println(manga.Name)
	case "download":
		var (
			chapter  *backends.Chapter
			chapters []*backends.Chapter
		)

		if manga == nil {
			fmt.Println("Error: you must select a manga before")
			return
		}

		chapters, err = d.Backend.Chapters(manga)
		if err != nil {
			panic(err)
		}
		chapter = chapters[0]
		fmt.Println(" :: Chapter =>", chapter)

		err = d.Download(manga, chapter, "./mangas")
		if err != nil {
			panic(err)
		}
	}
}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "backends", Description: "List available backends"},
		{Text: "use", Description: "Select backend to use"},
		{Text: "search", Description: "Search for a manga on selected backend"},
		{Text: "manga", Description: "Select manga with index"},
		{Text: "download", Description: "Download selected manga"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func main() {
	p = prompt.New(executor, completer, prompt.OptionPrefix(">>> "), prompt.OptionTitle("katago"))
	p.Run()
}
