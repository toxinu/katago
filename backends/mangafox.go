package backends

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"golang.org/x/net/html"
)

const (
	MangaFoxBaseURL                    = "http://mangafox.la"
	MangaFoxHTMLSelectorMangaName      = "#series_info div.cover img"
	MangaFoxHTMLSelectorMangaChapters1 = "#chapters ul.chlist li h3 a"
	MangaFoxHTMLSelectorMangaChapters2 = "#chapters ul.chlist li h4 a"
	MangaFoxHTMLSelectorChapterPages   = "#top_center_bar div.r option"
	MangaFoxHTMLSelectorPageImage      = "#image"
)

var (
	MangaFoxRegexpIdentifyManga   = regexp.MustCompile("^/manga/[0-9a-z_]+/?$")
	MangaFoxRegexpIdentifyChapter = regexp.MustCompile("^/manga/[0-9a-z_]+/.+$")
	MangaFoxRegexpChapterName     = regexp.MustCompile("^.*/c(\\d+(\\.\\d+)?).*$")
	MangaFoxRegexpPageBaseUrlPath = regexp.MustCompile("/?(\\d+\\.html)?$")
)

// MangaFox is MangaFox backend
type MangaFox struct {
	Downloader *Downloader
}

// Search implements Backend interface
func (b *MangaFox) Search(term string) ([]*Manga, error) {
	var (
		err             error
		body            []byte
		resp            *http.Response
		responseResults [][]string
		results         []*Manga
		manga           *Manga
		url             *url.URL
	)
	url, err = url.Parse(fmt.Sprintf("%s/ajax/search.php?term=%s", MangaFoxBaseURL, term))
	if err != nil {
		return nil, err
	}

	resp, err = b.Downloader.HTTPGet(url, []int{200})
	if err != nil {
		return nil, err
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	responseResults = make([][]string, 0)

	err = json.Unmarshal(body, &responseResults)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(responseResults); i++ {
		url, err = url.Parse(fmt.Sprintf("%s/manga/%s", MangaFoxBaseURL, responseResults[i][2]))
		if err != nil {
			return nil, err
		}
		manga = &Manga{
			ID:     responseResults[i][0],
			Name:   responseResults[i][1],
			Slug:   responseResults[i][2],
			Genre:  responseResults[i][3],
			Author: responseResults[i][4],
			URL:    url}
		results = append(results, manga)
	}

	return results, nil
}

// Chapters implements Backend interface
func (b *MangaFox) Chapters(manga *Manga) ([]*Chapter, error) {
	doc, err := b.Downloader.HTTPGetDocument(manga.URL, []int{200})
	if err != nil {
		return nil, err
	}

	linkNodes := make([]*html.Node, 0)
	linkNodes = append(linkNodes, doc.Find(MangaFoxHTMLSelectorMangaChapters1).Nodes...)
	linkNodes = append(linkNodes, doc.Find(MangaFoxHTMLSelectorMangaChapters2).Nodes...)

	chapters := make([]*Chapter, 0, len(linkNodes))
	for _, linkNode := range linkNodes {
		chapterURL, err := url.Parse(fmt.Sprintf("http:%s", htmlGetNodeAttribute(linkNode, "href")))
		if err != nil {
			return nil, err
		}

		chapterName, err := htmlGetNodeText(linkNode)
		if err != nil {
			return nil, err
		}

		chapter := &Chapter{URL: chapterURL, Name: chapterName}
		chapters = append(chapters, chapter)
	}
	chapters = chapterSliceReverse(chapters)

	return chapters, err
}

func (b *MangaFox) Pages(chapter *Chapter) ([]*Page, error) {
	doc, err := b.Downloader.HTTPGetDocument(chapter.URL, []int{200})
	if err != nil {
		return nil, err
	}

	basePageUrl := urlCopy(chapter.URL)
	basePageUrl.Path = MangaFoxRegexpPageBaseUrlPath.ReplaceAllString(basePageUrl.Path, "")

	optionNodes := doc.Find(MangaFoxHTMLSelectorChapterPages).Nodes
	pages := make([]*Page, 0, len(optionNodes))
	for _, optionNode := range optionNodes {
		pageNumberString := htmlGetNodeAttribute(optionNode, "value")
		pageNumber, err := strconv.Atoi(pageNumberString)
		if err != nil {
			return nil, err
		}

		if pageNumber <= 0 {
			continue
		}

		pageURL := urlCopy(basePageUrl)
		pageURL.Path += fmt.Sprintf("/%d.html", pageNumber)

		page := &Page{URL: pageURL}
		pages = append(pages, page)
	}

	return pages, nil
}

func (b *MangaFox) PageImageURL(page *Page) (*url.URL, error) {
	doc, err := b.Downloader.HTTPGetDocument(page.URL, []int{200})
	if err != nil {
		return nil, err
	}

	imgNodes := doc.Find(MangaFoxHTMLSelectorPageImage).Nodes
	if len(imgNodes) != 1 {
		return nil, fmt.Errorf("html node '%s' (page image url) not found in '%s'", MangaFoxHTMLSelectorPageImage, page.URL)
	}

	imgNode := imgNodes[0]

	imageURL, err := url.Parse(htmlGetNodeAttribute(imgNode, "src"))
	if err != nil {
		return nil, err
	}

	return imageURL, nil
}
