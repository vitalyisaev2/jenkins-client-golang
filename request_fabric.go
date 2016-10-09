package jenkins

import (
	"fmt"
	"io"
	"net/http"
)

type jenkinsAPIRequest struct {
	method      string
	route       string
	format      jenkinsAPIFormat
	body        io.Reader
	queryParams map[string]string
}

type jenkinsAPIFormat uint

const (
	jenkinsAPIFormatJSON jenkinsAPIFormat = iota
	jenkinsAPIFormatXML
)

type requestFabric struct {
	baseURL  string
	username string
	password string
}

func (rf *requestFabric) newURLString(route string, format jenkinsAPIFormat) string {
	var URL string
	switch format {
	case jenkinsAPIFormatXML:
		URL = fmt.Sprintf("%s%s", rf.baseURL, route)
	case jenkinsAPIFormatJSON:
		URL = fmt.Sprintf("%s%s/api/json", rf.baseURL, route)
	}
	return URL
}

func (rf *requestFabric) newHTTPRequest(apiRequest *jenkinsAPIRequest) (*http.Request, error) {
	// Create URL base
	URL := rf.newURLString(apiRequest.route, apiRequest.format)

	httpRequest, err := http.NewRequest(apiRequest.method, URL, apiRequest.body)
	if err != nil {
		return nil, err
	}

	// Build query params
	if apiRequest.queryParams != nil {
		query := httpRequest.URL.Query()
		for key, value := range apiRequest.queryParams {
			query.Add(key, value)
		}
		httpRequest.URL.RawQuery = query.Encode()
	}

	httpRequest.SetBasicAuth(rf.username, rf.password)
	return httpRequest, nil
}

func (rf *requestFabric) newJSONRequest(apiRequest *jenkinsAPIRequest) (*http.Request, error) {
	httpRequest, err := rf.newHTTPRequest(apiRequest)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Add("Content-Type", "application/json")
	return httpRequest, nil
}

func (rf *requestFabric) newXMLRequest(apiRequest *jenkinsAPIRequest) (*http.Request, error) {
	httpRequest, err := rf.newHTTPRequest(apiRequest)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Add("Content-Type", "application/xml")
	return httpRequest, nil
}
