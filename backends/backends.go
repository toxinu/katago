package backends

import (
	"net/url"

	"github.com/toxinu/katago/client"
)

// Manga represents a manga
type Manga struct {
	ID     string
	Name   string
	Slug   string
	Author string
	Genre  string
	URL    *url.URL
}

// Chapter represents a manga chapter
type Chapter struct {
	Name string
	URL  *url.URL
}

// Page represents a chapter page
type Page struct {
	URL *url.URL
}

// Backend interface represents a service to download manga
type Backend interface {
	Name() string
	Search(string) ([]*Manga, error)
	Chapters(*Manga) ([]*Chapter, error)
	Pages(*Chapter) ([]*Page, error)
	PageImageURL(*Page) (*url.URL, error)
}

// Backends is declared backends
var Backends map[string]Backend

// Initialize initialize every backends
func Initialize(c *client.Client) {
	Backends = map[string]Backend{
		"mangafox": &MangaFox{Client: c},
	}
}
