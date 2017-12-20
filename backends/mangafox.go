package backends

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/toxinu/katago/client"

	"golang.org/x/net/html"
)

const (
	// MangaFoxBaseURL is base URL for MangaFox backend
	MangaFoxBaseURL = "http://mangafox.la"
	// MangaFoxHTMLSelectorMangaName is manga name selector
	MangaFoxHTMLSelectorMangaName = "#series_info div.cover img"
	// MangaFoxHTMLSelectorMangaChapters1 is manga chapters selector
	MangaFoxHTMLSelectorMangaChapters1 = "#chapters ul.chlist li h3 a"
	// MangaFoxHTMLSelectorMangaChapters2 is manga chapters selector
	MangaFoxHTMLSelectorMangaChapters2 = "#chapters ul.chlist li h4 a"
	// MangaFoxHTMLSelectorChapterPages is chapter pages selector
	MangaFoxHTMLSelectorChapterPages = "#top_center_bar div.r option"
	// MangaFoxHTMLSelectorPageImage is page image selector
	MangaFoxHTMLSelectorPageImage = "#image"
)

var (
	// MangaFoxRegexpPageBaseURLPath is page URL regexp
	MangaFoxRegexpPageBaseURLPath = regexp.MustCompile("/?(\\d+\\.html)?$")
)

// MangaFox is MangaFox backend
type MangaFox struct {
	Client *client.Client
}

// Name implements Backend interface
func (*MangaFox) Name() string {
	return "MangaFox"
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

	resp, err = b.Client.Get(url, []int{200})
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
	doc, err := b.Client.GetDocument(manga.URL, []int{200})
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

// Pages implements Backend interface
func (b *MangaFox) Pages(chapter *Chapter) ([]*Page, error) {
	doc, err := b.Client.GetDocument(chapter.URL, []int{200})
	if err != nil {
		return nil, err
	}

	basePageURL := urlCopy(chapter.URL)
	basePageURL.Path = MangaFoxRegexpPageBaseURLPath.ReplaceAllString(basePageURL.Path, "")

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

		pageURL := urlCopy(basePageURL)
		pageURL.Path += fmt.Sprintf("/%d.html", pageNumber)

		page := &Page{URL: pageURL}
		pages = append(pages, page)
	}

	return pages, nil
}

// PageImageURL implements Backend interface
func (b *MangaFox) PageImageURL(page *Page) (*url.URL, error) {
	doc, err := b.Client.GetDocument(page.URL, []int{200})
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
