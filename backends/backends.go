package backends

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

var (
	regexpImageContentType, _ = regexp.Compile("^image/(.+)$")
)

type Manga struct {
	ID     string
	Name   string
	Slug   string
	Author string
	Genre  string
	URL    *url.URL
}

type Chapter struct {
	Name string
	URL  *url.URL
}

type Page struct {
	URL *url.URL
}

type Backend interface {
	Search(string) ([]*Manga, error)
	Chapters(*Manga) ([]*Chapter, error)
	Pages(*Chapter) ([]*Page, error)
	PageImageURL(*Page) (*url.URL, error)
}

type Downloader struct {
	Backend         Backend
	ParallelChapter int
	ParallelPage    int
	HTTPRetry       int
}

func NewDownloader(backendName string) *Downloader {
	backend, ok := Backends[backendName]
	if !ok {
		panic("backend not found")
	}
	return &Downloader{
		Backend:         backend,
		ParallelChapter: 5,
		ParallelPage:    5,
		HTTPRetry:       10,
	}
}

func (d *Downloader) HTTPGet(u *url.URL, successCodes []int) (*http.Response, error) {
	var (
		err  error
		req  *http.Request
		resp *http.Response
	)

	req, err = http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/31.0.1650.4 Safari/537.36")

	for i := 0; i < 10; i++ {
		resp, err = http.DefaultClient.Do(req)
		if err == nil && (len(successCodes) == 1 && (successCodes[0] == 0 || contains(successCodes, resp.StatusCode))) {
			time.Sleep(500 * time.Millisecond)
			return resp, err
		}
	}

	return nil, errors.New("Success code never reached")
}

func (d *Downloader) HTTPGetBody(u *url.URL, successCodes []int) (*html.Node, error) {
	resp, err := d.HTTPGet(u, successCodes)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	node, err := html.Parse(resp.Body)
	return node, err
}

func (d *Downloader) HTTPGetDocument(u *url.URL, successCodes []int) (*goquery.Document, error) {
	node, err := d.HTTPGetBody(u, successCodes)
	if err != nil {
		return nil, err
	}

	return goquery.NewDocumentFromNode(node), err
}

func (d *Downloader) Download(manga *Manga, chapter *Chapter, output string) error {
	var (
		waitGroup sync.WaitGroup
	)

	output = path.Join(output, manga.Name, chapter.Name)
	fmt.Println(" :: Output  => %s", output)

	pages, err := d.Backend.Pages(chapter)
	if err != nil {
		return err
	}

	type pageTask struct {
		page  *Page
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
				result <- d.DownloagePage(chapterPageTask.page, chapterPageTask.index, output)
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

func (d *Downloader) DownloagePage(page *Page, index int, output string) error {
	pagePath := path.Join(output, strconv.Itoa(index))

	imageURL, err := d.Backend.PageImageURL(page)
	if err != nil {
		return err
	}

	resp, err := d.HTTPGet(imageURL, make([]int, 1))
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

var Backends = map[string]Backend{
	"mangafox": &MangaFox{},
}
