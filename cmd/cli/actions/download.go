package actions

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cheggaaa/pb"
	"github.com/toxinu/katago/backends"
	"github.com/toxinu/katago/downloader"
)

// Download represents a download cli action
type Download struct{}

// Run implements Action interface
func (a *Download) Run(ctx context.Context, parameters []string) context.Context {
	var (
		err                error
		d                  *downloader.Downloader
		index              int
		manga              *backends.Manga
		chapter            *backends.Chapter
		chapters           []*backends.Chapter
		chaptersToDownload []*backends.Chapter
		indexes            []int
		start              int
		end                int
	)

	if FromContext(ctx, "manga") == nil {
		PrintError(errors.New("you must select a manga before"))
		return ctx
	}

	for _, i := range parameters {
		if strings.Contains(i, "-") {
			splittedIndex := strings.Split(i, "-")
			if len(splittedIndex) > 2 {
				PrintError(fmt.Errorf("invalid chapter range: \"%s\"", i))
				return ctx
			}

			start, err = strconv.Atoi(splittedIndex[0])
			if err != nil {
				PrintError(fmt.Errorf("invalid start chapter index (must be integer): \"%s\"", i))
				return ctx
			}

			end, err = strconv.Atoi(splittedIndex[1])
			if err != nil {
				PrintError(fmt.Errorf("invalid end chapter index (must be integer): \"%s\"", i))
				return ctx
			}

			if start > end {
				PrintError(fmt.Errorf("invalid chapters range: \"%s\"", i))
				return ctx
			}
			for y := start; y <= end; y++ {
				indexes = append(indexes, y)
			}
		} else {
			index, err = strconv.Atoi(i)
			if err != nil {
				PrintError(fmt.Errorf("invalid chapter index (must be integer): \"%s\"", i))
				return ctx
			}
			indexes = []int{index}
		}
	}

	fmt.Println(" => Chapters to download:", indexes)

	manga = FromContext(ctx, "manga").(*backends.Manga)
	d = FromContext(ctx, "downloader").(*downloader.Downloader)

	chapters, err = d.Backend.Chapters(manga)
	if err != nil || len(chapters) == 0 {
		PrintError(errors.New("cannot retrieve chapters"))
		return ctx
	}

	bar := pb.StartNew(len(indexes))

	for _, i := range indexes {
		if i >= len(chapters) {
			PrintError(fmt.Errorf("chapter \"%d\" is not available", i))
			return ctx
		}

		chapter = chapters[i]

		chaptersToDownload = append(chaptersToDownload, chapter)
	}

	results := make(chan error)
	d.Download(manga, chaptersToDownload, "./mangas", results)

	for err := range results {
		if err != nil {
			PrintError(err)
		}
		bar.Increment()
	}

	bar.Finish()
	fmt.Printf("\nDone! :-)\n")

	return ctx
}

// Tips implements action interface
func (*Download) Tips() {
}

// Help implements action interface
func (*Download) Help() {}
