package src

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
)

// HTTPSource is the structure to get data
type HTTPSource struct {
	URI *url.URL
}

// GetData gets data from a http resource
func (h HTTPSource) GetData() ([]byte, error) {
	client := http.Client{}
	resp, err := client.Get(h.URI.String())
	if err != nil {
		return nil, fmt.Errorf("unable to get webpage %s [%+v]", h.URI.String(), err)
	}
	defer resp.Body.Close()
	b := &bytes.Buffer{}
	_, err = b.ReadFrom(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response from %s [%+v]", h.URI.String(), err)
	}
	return b.Bytes(), nil
}
