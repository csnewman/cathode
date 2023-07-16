package mediaserver

type MediaType string

const (
	MediaTypeTv    MediaType = "tv"
	MediaTypeMovie MediaType = "movie"
)

type SearchResult struct {
	Type         MediaType
	OriginalName string
	OriginalLang string
	Overview     string
	Popularity   float32
}

type Index interface {
	Search(name string) ([]*SearchResult, error)
}
