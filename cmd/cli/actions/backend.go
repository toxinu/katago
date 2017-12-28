package actions

import (
	"context"
	"errors"

	"github.com/toxinu/katago/downloader"
)

// Backend represents backend cli action
type Backend struct{}

// Run implements Action interface
func (a *Backend) Run(ctx context.Context, parameters []string) context.Context {
	var (
		d   *downloader.Downloader
		err error
	)

	if len(parameters) == 0 {
		PrintError(errors.New("backend name needed"))
		return ctx
	}

	backend := parameters[0]

	d, err = downloader.NewDownloader(backend)
	if err != nil {
		PrintError(err)
		return ctx
	}
	ctx = ToContext(ctx, "downloader", d)

	return ctx
}

// Tips implements Action interface
func (*Backend) Tips() {}

// Help implements Action interface
func (*Backend) Help() {}
