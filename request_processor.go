package jenkins

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
)

type requestProcessor struct {
	client *http.Client
	fabric *requestFabric
}

// HTTP method GET
func (processor *requestProcessor) getJSON(apiRequest *jenkinsAPIRequest, receiver Result) error {
	var err error
	var httpRequest *http.Request
	var httpResponse *http.Response

	httpRequest, err = processor.fabric.newJSONRequest(apiRequest)
	if err != nil {
		return err
	}

	httpResponse, err = processor.client.Do(httpRequest)
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		location, _ := httpResponse.Location()
		return fmt.Errorf("%v: %s", location, httpResponse.Status)
	}

	err = json.NewDecoder(httpResponse.Body).Decode(receiver)
	return err
}

//HTTP method Post
//func (processor *requestProcessor) postXML(apiRequest *jenkinsAPIRequest, receiver Result) error {
//var err error
//var httpRequest *http.Request
//var httpResponse *http.Response

//httpRequest, err = processor.fabric.newJSONRequest(apiRequest)
//if err != nil {
//return err
//}
//}

func newRequestProcessor(baseURL string, username string, password string) (*requestProcessor, error) {

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

	return &requestProcessor{client, fabric}, nil
}
