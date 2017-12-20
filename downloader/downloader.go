package downloader

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"

	"github.com/toxinu/katago/backends"
	"github.com/toxinu/katago/client"
)

var (
	regexpImageContentType, _ = regexp.Compile("^image/(.+)$")
)

// Downloader manage HTTP requests to download manga
type Downloader struct {
	Backend         backends.Backend
	Client          *client.Client
	ParallelChapter int
	ParallelPage    int
}

// NewDownloader return a Downloader
func NewDownloader(backendName string) *Downloader {
	c := &client.Client{Retry: 10}
	backends.Initialize(c)

	b, ok := backends.Backends[backendName]
	if !ok {
		panic("backend not found")
	}
	return &Downloader{
		Backend:         b,
		Client:          c,
		ParallelChapter: 5,
		ParallelPage:    5,
	}
}

// Download retrieve a manga's chapter
func (d *Downloader) Download(manga *backends.Manga, chapter *backends.Chapter, output string) error {
	var (
		waitGroup sync.WaitGroup
	)

	output = path.Join(output, manga.Name, chapter.Name)
	fmt.Println(" :: Output  =>", output)

	pages, err := d.Backend.Pages(chapter)
	if err != nil {
		return err
	}

	type pageTask struct {
		page  *backends.Page
		index int
	}

	tasks := make(chan *pageTask)
	go func() {
		for index, page := range pages {
			tasks <- &pageTask{
				page:  page,
				index: index + 1,
			}
		}
		close(tasks)
	}()

	waitGroup.Add(d.ParallelPage)
	result := make(chan error)
	for i := 0; i < d.ParallelPage; i++ {
		go func() {
			for chapterPageTask := range tasks {
				result <- d.DownloadPage(chapterPageTask.page, chapterPageTask.index, output)
			}
			waitGroup.Done()
		}()
	}

	go func() {
		waitGroup.Wait()
		close(result)
	}()

	for err := range result {
		if err != nil {
			return err
		}
	}

	return nil
}

// DownloadPage retrieve a Manga Page
func (d *Downloader) DownloadPage(page *backends.Page, index int, output string) error {
	pagePath := path.Join(output, strconv.Itoa(index))

	imageURL, err := d.Backend.PageImageURL(page)
	if err != nil {
		return err
	}

	resp, err := d.Client.Get(imageURL, make([]int, 1))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var extension string
	if len(extension) == 0 {
		contentType := resp.Header.Get("content-type")
		if len(contentType) > 0 {
			matches := regexpImageContentType.FindStringSubmatch(contentType)
			if matches != nil {
				extension = matches[1]
			}
		}
	}
	if len(extension) > 0 {
		if extension == "jpeg" {
			extension = "jpg"
		}
		pagePath += "." + extension
	}

	err = os.MkdirAll(filepath.Dir(pagePath), 0755)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(pagePath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
