package request

import (
	"fmt"
	"io"
	"net/http"
)

type JenkinsAPIRequest struct {
	Method      string
	Route       string
	Format      JenkinsAPIFormat
	Body        io.Reader
	QueryParams map[string]string
}

type JenkinsAPIFormat uint

const (
	JenkinsAPIFormatJSON JenkinsAPIFormat = iota
	JenkinsAPIFormatXML
)

type fabric struct {
	baseURL  string
	username string
	password string
}

func (rf *fabric) newURLString(route string, format JenkinsAPIFormat) string {
	var URL string
	switch format {
	case JenkinsAPIFormatXML:
		URL = fmt.Sprintf("%s%s", rf.baseURL, route)
	case JenkinsAPIFormatJSON:
		URL = fmt.Sprintf("%s%s/api/json", rf.baseURL, route)
	}
	return URL
}

// Creates arbitrary HTTP Request
func (rf *fabric) newHTTPRequest(apiRequest *JenkinsAPIRequest) (*http.Request, error) {
	// Create URL base
	URL := rf.newURLString(apiRequest.Route, apiRequest.Format)

	httpRequest, err := http.NewRequest(apiRequest.Method, URL, apiRequest.Body)
	if err != nil {
		return nil, err
	}

	// Build query params
	if apiRequest.QueryParams != nil {
		query := httpRequest.URL.Query()
		for key, value := range apiRequest.QueryParams {
			query.Add(key, value)
		}
		httpRequest.URL.RawQuery = query.Encode()
	}

	httpRequest.SetBasicAuth(rf.username, rf.password)
	return httpRequest, nil
}

// Creates new HTTP Request used for crumb generation
func (rf *fabric) newCrumbRequest() (*http.Request, error) {
	URL := rf.newURLString("/crumbIssuer", JenkinsAPIFormatJSON)
	httpRequest, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	httpRequest.SetBasicAuth(rf.username, rf.password)
	return httpRequest, nil
}
