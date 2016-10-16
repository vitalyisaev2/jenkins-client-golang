package request

import (
	"fmt"
	"io"
	"net/http"
)

// JenkinsAPIRequest instances are passed to requester from
// exported high-level functions
type JenkinsAPIRequest struct {
	Method      string
	Route       string
	Body        io.Reader
	QueryParams map[string]string
	Format      JenkinsAPIFormat
	DumpMethod  ResponseDumpMethod
}

// JenkinsAPIFormat turns on JSON or XML responses from Jenkins API
type JenkinsAPIFormat uint

const (
	// JenkinsAPIFormatJSON appends /api/json to request routes
	JenkinsAPIFormatJSON JenkinsAPIFormat = iota
	// JenkinsAPIFormatXML appends /api/xml to request routes
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
		URL = fmt.Sprintf("%s%s/api/xml", rf.baseURL, route)
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
