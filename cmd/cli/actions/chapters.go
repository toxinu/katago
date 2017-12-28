package actions

import (
	"context"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/toxinu/katago/backends"
	"github.com/toxinu/katago/cmd/cli/colors"
	"github.com/toxinu/katago/downloader"
)

// Chapters represents chapters cli action
type Chapters struct{}

// Run implements Action interface
func (a *Chapters) Run(ctx context.Context, parameters []string) context.Context {
	var (
		err      error
		d        *downloader.Downloader
		manga    *backends.Manga
		chapters []*backends.Chapter
		columns  = 2
	)

	if FromContext(ctx, "manga") == nil {
		PrintError(errors.New("you must select a manga before"))
		return ctx
	}

	manga = FromContext(ctx, "manga").(*backends.Manga)
	d = FromContext(ctx, "downloader").(*downloader.Downloader)

	chapters, err = d.Backend.Chapters(manga)
	if err != nil {
		PrintError(errors.New("cannot retrieve chapters"))
		return ctx
	}

	if len(chapters) == 0 {
		fmt.Println("No chapters found")
		return ctx
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.FilterHTML)

	for index := 0; index-columns <= len(chapters); index = index + columns {
		start, end := index-columns, index

		if start < 0 {
			start = 0
		}

		if end >= len(chapters) {
			end = len(chapters)
		}

		c := chapters[start:end]
		for i := 0; i < len(c); i++ {
			fmt.Fprintf(w, "%s%d%s.\t %s\t |", colors.Bright, start+i, colors.Reset, c[i].Name)
		}
		fmt.Fprintf(w, "\n")
	}

	w.Flush()

	return ctx
}

// Tips implements action interface
func (*Chapters) Tips() {
}

// Help implements action interface
func (*Chapters) Help() {}
