package backends

type Manga struct {
	Name string
}

type Backend interface {
	Name() string
}

type Downloader struct {
	BackendName string
}

func (d *Downloader) Search(term string) []string {
	return d.GetBackend().Search(term)
}

func (d *Downloader) GetBackend() Backend {
	b, ok := Backends[d.BackendName]
	if !ok {
		panic("backend not found")
	}
	return b
}

func NewDownloader(backend string) *Downloader {
	return &Downloader{BackendName: backend}
}

var Backends = map[string]Backend{
	"mangafox": MangaFox{},
}
