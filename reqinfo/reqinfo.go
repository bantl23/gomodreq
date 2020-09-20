package reqinfo

import (
	"fmt"
	"net/url"

	"github.com/bantl23/gomodreq/src"
	"gopkg.in/yaml.v2"
)

// ReqInfo contains the golang modules version requirements
type ReqInfo struct {
	Required map[string]string   `yaml:"required,omitempty"`
	Banned   map[string][]string `yaml:"banned,omitempty"`
}

// GetReqInfo returns requirements info from gomodreq file
func GetReqInfo(loc string) (*ReqInfo, error) {
	req := new(ReqInfo)

	uri, err := url.ParseRequestURI(loc)
	if err != nil {
		return nil, fmt.Errorf("unable to parse uri %s [%+v]", uri, err)
	}

	source := src.Source{URI: uri}
	data, err := source.GetData()
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, req)
	if err != nil {
		return nil, fmt.Errorf("unable to parse requirements file %s [%+v]", uri, err)
	}

	return req, nil
}
