package src

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
)

// FileSource is the structure used to get data
type FileSource struct {
	URI *url.URL
}

// GetData gets data from a file
func (f FileSource) GetData() ([]byte, error) {
	_, err := os.Stat(f.URI.Path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist %s [%+v]", f.URI.Path, err)
	}
	data, err := ioutil.ReadFile(f.URI.Path)
	if err != nil {
		return nil, fmt.Errorf("problem reading file %s [%+v]", f.URI.Path, err)
	}
	return data, nil
}
