package src

import (
	"fmt"
	"net/url"
)

// SourceI is the inteface to get data
type SourceI interface {
	GetData(*url.URL) ([]byte, error)
}

// Source is the structured used to get data
type Source struct {
	URI *url.URL
}

// GetData gets the data from requirements source
func (s Source) GetData() ([]byte, error) {
	switch s.URI.Scheme {
	case "file":
		return FileSource{URI: s.URI}.GetData()
	case "http", "https":
		return HTTPSource{URI: s.URI}.GetData()
	case "ssh":
		return SSHSource{URI: s.URI}.GetData()
	default:
		return nil, fmt.Errorf("Unsupported uri scheme: %s", s.URI.Scheme)
	}
}
