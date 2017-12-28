package actions

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/toxinu/katago/backends"
	"github.com/toxinu/katago/cmd/cli/colors"
	"github.com/toxinu/katago/downloader"
)

// Search represents a search cli action
type Search struct{}

// Run implements action interface
func (a *Search) Run(ctx context.Context, parameters []string) context.Context {
	var (
		d       *downloader.Downloader
		err     error
		results []*backends.Manga
	)

	d = FromContext(ctx, "downloader").(*downloader.Downloader)
	results, err = d.Backend.Search(strings.Join(parameters, " "))
	if err != nil {
		PrintError(err)
		return ctx
	}

	ctx = ToContext(ctx, "results", results)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	for index, result := range results {
		fmt.Fprintf(w, "%s%d%s\t | %s\t | %s\n", colors.Bright, index, colors.Reset, result.Name, result.Author)
	}
	w.Flush()

	return ctx
}

// Tips implements action interface
func (*Search) Tips() {
	fmt.Println("\n => Tips: to select a manga, use `manga <index>`")
}

// Help implements action interface
func (*Search) Help() {}
