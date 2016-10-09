package jenkins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
)

type requestProcessor struct {
	client *http.Client
	fabric *requestFabric
	debug  bool
}

func (processor *requestProcessor) getJSON(apiRequest *jenkinsAPIRequest, receiver Result) error {
	var err error
	var httpRequest *http.Request

	httpRequest, err = processor.fabric.newHTTPRequest(apiRequest)
	if err != nil {
		return err
	}
	httpRequest.Header.Add("Content-Type", "application/json")

	return processor.processRequest(httpRequest, receiver)
}

func (processor *requestProcessor) postXML(apiRequest *jenkinsAPIRequest, receiver Result) error {
	var err error
	var httpRequest *http.Request

	httpRequest, err = processor.fabric.newHTTPRequest(apiRequest)
	if err != nil {
		return err
	}

	err = processor.setCrumbs(httpRequest)
	if err != nil {
		return err
	}

	httpRequest.Header.Add("Content-Type", "application/xml")

	return processor.processRequest(httpRequest, receiver)
}

// Make HTTP Request match Jenkins CSRF protection requirements (enabled by default in 2.x)
func (processor *requestProcessor) setCrumbs(httpRequest *http.Request) error {
	var err error
	var crumbRequest *http.Request

	crumbRequest, err = processor.fabric.newCrumbRequest()
	if err != nil {
		return err
	}
	receiver := make(map[string]string)

	err = processor.processRequest(crumbRequest, &receiver)
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

// Enqueue HTTP request to client
func (processor *requestProcessor) processRequest(httpRequest *http.Request, receiver Result) error {
	var err error
	var httpResponse *http.Response

	httpResponse, err = processor.client.Do(httpRequest)
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()

	httpRequestURL := httpRequest.URL.String()
	if httpResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("%v: %s", httpRequestURL, httpResponse.Status)
	}

	switch processor.debug {
	case true:
		{
			dumpedBody, _ := ioutil.ReadAll(httpResponse.Body)
			dumpedBodyReader := bytes.NewBuffer(dumpedBody)
			fmt.Printf("URL: %s ResponseBody: %s\n", httpRequestURL, string(dumpedBody))
			switch receiver {
			case nil:
				return nil
			default:
				err = json.NewDecoder(dumpedBodyReader).Decode(receiver)
			}
		}
	case false:
		{
			switch receiver {
			case nil:
				return nil
			default:
				err = json.NewDecoder(httpResponse.Body).Decode(receiver)
			}
		}

	}
	return err
}

// Creates new wrapper around standard http.Client
func newRequestProcessor(baseURL string, username string, password string, debug bool) (*requestProcessor, error) {

	var (
		err       error
		cookieJar *cookiejar.Jar
		transport *http.Transport
		client    *http.Client
		fabric    *requestFabric
	)

	// Build custom http/client
	transport = &http.Transport{
		MaxIdleConnsPerHost: 16,
	}

	// Construct cookie storage
	cookieJar, err = cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client = &http.Client{
		Transport: transport,
		Jar:       cookieJar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			req.SetBasicAuth(username, password)
			return nil
		},
	}

	// requestFabric creates various requests
	fabric = &requestFabric{baseURL, username, password}

	return &requestProcessor{client, fabric, debug}, nil
}
