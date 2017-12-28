package main

import (
	"context"
	"fmt"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/toxinu/katago/backends"
	"github.com/toxinu/katago/cmd/cli/actions"
	"github.com/toxinu/katago/downloader"
)

var ctx = context.Background()

// FromContext returns key's value from given Context
func FromContext(ctx context.Context, key string) interface{} {
	return ctx.Value(actions.Key(key))
}

// ToContext sets given value to Context's key
func ToContext(ctx context.Context, key string, value interface{}) context.Context {
	return context.WithValue(ctx, actions.Key(key), value)
}

func executor(in string) {
	splitted := strings.Fields(in)
	if len(splitted) == 0 {
		return
	}
	action := splitted[0]
	parameters := splitted[1:len(splitted)]

	ctx = actions.Run(ctx, action, parameters)
}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "backends", Description: "List available backends"},
		{Text: "backend", Description: "Select backend to use"},
		{Text: "search", Description: "Search for a manga on selected backend"},
		{Text: "manga", Description: "Select manga with index"},
		{Text: "download", Description: "Download selected manga"},
		{Text: "chapters", Description: "List selected manga chapters"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func main() {
	ctx = ToContext(ctx, "backend", "mangafox")

	d, err := downloader.NewDownloader(FromContext(ctx, "backend").(string))
	if err != nil {
		panic(err)
	}

	ctx = ToContext(ctx, "downloader", d)
	ctx = ToContext(ctx, "manga", nil)
	ctx = ToContext(ctx, "prompt", prompt.New(executor, completer, prompt.OptionPrefix(">>> "), prompt.OptionTitle("katago")))
	ctx = ToContext(ctx, "results", []*backends.Manga{})

	fmt.Println("Welcome,")
	fmt.Printf("I have already selected \"%s\" backend for you :)\n\n", FromContext(ctx, "backend"))
	fmt.Println("You can now `search` for a manga.")

	FromContext(ctx, "prompt").(*prompt.Prompt).Run()
}
