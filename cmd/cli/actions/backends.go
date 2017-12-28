package actions

import (
	"context"
	"fmt"

	"github.com/toxinu/katago/backends"
	"github.com/toxinu/katago/downloader"
)

// Backends represents backends cli action
type Backends struct{}

// Run implements Action interface
func (a *Backends) Run(ctx context.Context, parameters []string) context.Context {
	d := FromContext(ctx, "downloader").(*downloader.Downloader)

	for slug, backend := range backends.Backends {
		if backend.Name() == d.Backend.Name() {
			fmt.Println(slug, "(enabled)")
		} else {
			fmt.Println(slug)
		}
	}
	return ctx
}

// Tips implements Action interface
func (*Backends) Tips() {
	fmt.Println("\n => Tips: to select a backend, use `backend <name>`")
}

// Help implements Action interface
func (*Backends) Help() {
	fmt.Println("Show available backends")
}
