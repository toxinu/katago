package downloader

import (
	"io/ioutil"
	"net/http"
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

// Downloader downloads manga
type Downloader struct {
	Backend         backends.Backend
	Client          *client.Client
	ParallelChapter int
	ParallelPage    int
}

// NewDownloader returns a Downloader
func NewDownloader(backendName string) (*Downloader, error) {
	c := &client.Client{Retry: 10}

	backends.Initialize(c)
	b, err := backends.Get(backendName)
	if err != nil {
		return nil, err
	}

	return &Downloader{
		Backend:         b,
		Client:          c,
		ParallelChapter: 5,
		ParallelPage:    5,
	}, nil
}

// Download retrieves a manga's chapters
func (d *Downloader) Download(manga *backends.Manga, chapters []*backends.Chapter, output string, results chan<- error) {
	var waitGroup sync.WaitGroup

	type chapterTask struct {
		manga   *backends.Manga
		chapter *backends.Chapter
	}

	tasks := make(chan *chapterTask)
	go func() {
		for _, chapter := range chapters {
			tasks <- &chapterTask{
				manga:   manga,
				chapter: chapter,
			}
		}
		close(tasks)
	}()

	waitGroup.Add(d.ParallelChapter)
	for i := 0; i < d.ParallelChapter; i++ {
		go func() {
			for mangaChapterTask := range tasks {
				results <- d.DownloadChapter(mangaChapterTask.manga, mangaChapterTask.chapter, output)
			}
			waitGroup.Done()
		}()
	}

	go func() {
		waitGroup.Wait()
		close(results)
	}()
}

// DownloadChapter retrieves a manga's chapter
func (d *Downloader) DownloadChapter(manga *backends.Manga, chapter *backends.Chapter, output string) error {
	var (
		waitGroup sync.WaitGroup
	)

	output = path.Join(output, manga.Name, chapter.Name)

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
	var (
		err  error
		resp *http.Response
	)

	pagePath := path.Join(output, strconv.Itoa(index))

	for i := 0; i < 10; i++ {
		imageURL, err := d.Backend.PageImageURL(page)
		if err != nil {
			continue
		}

		// resp, err := d.Client.Get(imageURL, make([]int, 1))
		resp, err = d.Client.Get(imageURL, []int{200})
		if err != nil {
			continue
		} else {
			break
		}
	}

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
