package mediaserver

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
)

// TMDBAPIKey for themoviedb.org. This key has no value or permissions and does not need to be treated as a secret.
const TMDBAPIKey = "6af65a7bd8013744b97fae599a76bf4b" //nolint:gosec

var ErrUnexpectedFormat = errors.New("unexpected format")

type MovieDB struct {
	c *tmdb.Client
}

func NewMoveDB() *MovieDB {
	c, err := tmdb.Init(TMDBAPIKey)
	if err != nil {
		// No error is possible
		panic(err)
	}

	c.SetClientAutoRetry()

	// TODO: Allow user to customise
	proxyURL, err := url.Parse("http://127.0.0.1:8888")
	if err != nil {
		// No error is possible
		panic(err)
	}

	c.SetClientConfig(http.Client{
		Timeout: time.Second * 30,
		Transport: &http.Transport{
			MaxIdleConns:    10,
			IdleConnTimeout: 15 * time.Second,
			Proxy:           http.ProxyURL(proxyURL),
		},
	})

	return &MovieDB{
		c: c,
	}
}

func (m *MovieDB) Search(name string) ([]*SearchResult, error) {
	res, err := m.c.GetSearchMulti(name, map[string]string{})
	if err != nil {
		return nil, err
	}

	var mapped []*SearchResult

	for _, result := range res.Results {
		sr := &SearchResult{
			OriginalLang: result.OriginalLanguage,
			Overview:     result.Overview,
			Popularity:   result.Popularity,
		}

		switch result.MediaType {
		case "tv":
			sr.Type = MediaTypeTv
			sr.OriginalName = result.OriginalName
		case "movie":
			sr.Type = MediaTypeMovie
			sr.OriginalName = result.OriginalTitle
		case "person":
			// Unsupported
			continue
		default:
			return nil, fmt.Errorf("%w: unknown media type %s", ErrUnexpectedFormat, result.MediaType)
		}

		mapped = append(mapped, sr)
	}

	return mapped, nil
}
