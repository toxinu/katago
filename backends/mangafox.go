package backends

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

const (
	MangaFoxHtmlSelectorMangaName      = "#series_info div.cover img"
	MangaFoxHtmlSelectorMangaChapters1 = "#chapters ul.chlist li h3 a"
	MangaFoxHtmlSelectorMangaChapters2 = "#chapters ul.chlist li h4 a"
	MangaFoxHtmlSelectorChapterPages   = "#top_center_bar div.r option"
	MangaFoxHtmlSelectorPageImage      = "#image"
)

var (
	MangaFoxHosts = []string{
		"mangafox.me",
		"beta.mangafox.com",
	}

	MangaFoxRegexpIdentifyManga   = regexp.MustCompile("^/manga/[0-9a-z_]+/?$")
	MangaFoxRegexpIdentifyChapter = regexp.MustCompile("^/manga/[0-9a-z_]+/.+$")
	MangaFoxRegexpChapterName     = regexp.MustCompile("^.*/c(\\d+(\\.\\d+)?).*$")
	MangaFoxRegexpPageBaseUrlPath = regexp.MustCompile("/?(\\d+\\.html)?$")
)

type MangaFox struct {
	Downloader *Downloader
}

type MangaFoxSearchResult struct {
	Result []string
}

type MangaFoxSearchResults struct {
}

func (MangaFox) Search(term string) []string {
	resp, err := http.Get(fmt.Sprintf("http://mangafox.la/ajax/search.php?term=%s", term))
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		panic(readErr)
	}

	results := make([]string, 0)
	json.Unmarshal(body, &results)

	return results
}

func (MangaFox) Name() string {
	return "MangaFox"
}
