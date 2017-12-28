package actions

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/toxinu/katago/backends"
)

// Manga represents manga cli action
type Manga struct{}

// Run implements Action interface
func (a *Manga) Run(ctx context.Context, parameters []string) context.Context {
	var (
		err     error
		index   int
		manga   *backends.Manga
		results []*backends.Manga
	)

	results = FromContext(ctx, "results").([]*backends.Manga)

	if len(parameters) == 0 {
		PrintError(errors.New("manga index needed"))
		return ctx
	}

	index, err = strconv.Atoi(parameters[0])
	if err != nil {
		PrintError(errors.New("invalid index (must be integer)"))
		return ctx
	}

	if index >= len(results) {
		PrintError(errors.New("index out of range"))
		return ctx
	}

	manga = results[index]
	ctx = ToContext(ctx, "manga", manga)

	fmt.Printf("I confirm you that you just select this manga:\n\n")
	fmt.Println("Name:", manga.Name)
	fmt.Println("Author:", manga.Author)
	fmt.Println("Genre:", manga.Genre)

	return ctx
}

// Tips implements Action interface
func (*Manga) Tips() {}

// Help implements Action interface
func (*Manga) Help() {}
