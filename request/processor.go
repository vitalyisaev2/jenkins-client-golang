package request

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"

	"golang.org/x/net/context/ctxhttp"
)

// Processor wraps routines related to the HTTP layer of interaction with Jenkins API
type Processor interface {
	GetJSON(*JenkinsAPIRequest, interface{}) error
	Post(*JenkinsAPIRequest, interface{}) error
	PostXML(*JenkinsAPIRequest, interface{}) error
}

type defaultProcessor struct {
	client *http.Client
	fb     *fabric
	dm     *dumper
	debug  bool
}

func (p *defaultProcessor) GetJSON(apiRequest *JenkinsAPIRequest, receiver interface{}) error {
	httpRequest, err := p.fb.newHTTPRequest(apiRequest)
	if err != nil {
		return err
	}
	httpRequest.Header.Add("Content-Type", "application/json")
	return p.call(httpRequest, receiver, apiRequest.DumpMethod, true)
}

func (p *defaultProcessor) Post(apiRequest *JenkinsAPIRequest, receiver interface{}) error {
	httpRequest, err := p.fb.newHTTPRequest(apiRequest)
	if err != nil {
		return err
	}
	return p.call(httpRequest, receiver, apiRequest.DumpMethod, true)
}

func (p *defaultProcessor) PostXML(apiRequest *JenkinsAPIRequest, receiver interface{}) error {
	httpRequest, err := p.fb.newHTTPRequest(apiRequest)
	if err != nil {
		return err
	}
	httpRequest.Header.Add("Content-Type", "application/xml")
	return p.call(httpRequest, receiver, apiRequest.DumpMethod, true)
}

// Make HTTP Request match Jenkins CSRF protection requirements
// (enabled by default in 2.x)
func (p *defaultProcessor) setCrumbs(httpRequest *http.Request) error {
	var err error
	var crumbRequest *http.Request

	crumbRequest, err = p.fb.newCrumbRequest()
	if err != nil {
		return err
	}
	receiver := make(map[string]string)

	err = p.call(crumbRequest, &receiver, ResponseDumpDefaultJSON, false)
	if err != nil {
		return err
	}

	if _, ok := receiver["crumbRequestField"]; !ok {
		return fmt.Errorf("setCrumbs: %v has no field 'crumbRequestField'", receiver)
	}
	if _, ok := receiver["crumb"]; !ok {
		return fmt.Errorf("setCrumbs: %v has no field 'crumbRequestField'", receiver)
	}

	httpRequest.Header.Add(receiver["crumbRequestField"], receiver["crumb"])
	return nil
}

// Emit HTTP request to Jenkins endpoint and
func (p *defaultProcessor) call(
	req *http.Request,
	receiver interface{},
	dumpMethod ResponseDumpMethod,
	setCrumbs bool,
) error {

	if p.debug {
		fmt.Printf("Request URL: %s\n", req.URL)
	}

	// Set header preventing CSRF attacs if necessary
	if setCrumbs {
		if err := p.setCrumbs(req); err != nil {
			return err
		}
	}

	// Perform HTTP request
	resp, err := ctxhttp.Do(context.Background(), p.client, req)
	if err != nil {
		return err
	}

	return p.dm.dump(resp, receiver, dumpMethod)
}

// NewProcessor instantiates Processor - a wrapper for http.Client
// that aware about Jenkins features
func NewProcessor(
	baseURL string,
	username string,
	password string,
	debug bool,
) (Processor, error) {

	// Build custom http/client
	transport := &http.Transport{
		MaxIdleConnsPerHost: 16,
	}

	// Construct cookie storage
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: transport,
		Jar:       cookieJar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			req.SetBasicAuth(username, password)
			return nil
		},
	}

	// fabric creates various HTTP requests
	fb := &fabric{baseURL, username, password}

	// dumper deserializes HTTP responses to structs in various ways
	dm := &dumper{debug}

	return &defaultProcessor{client, fb, dm, debug}, nil
}
