package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vitalyisaev2/jenkins-client-golang/result"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
)

type Processor interface {
	GetJSON(*JenkinsAPIRequest, result.Result) error
	PostXML(*JenkinsAPIRequest, result.Result) error
}

type processorImpl struct {
	client *http.Client
	fb     *fabric
	debug  bool
}

func (processor *processorImpl) GetJSON(apiRequest *JenkinsAPIRequest, receiver result.Result) error {
	var err error
	var httpRequest *http.Request

	httpRequest, err = processor.fb.newHTTPRequest(apiRequest)
	if err != nil {
		return err
	}
	httpRequest.Header.Add("Content-Type", "application/json")

	return processor.processRequest(httpRequest, receiver)
}

func (processor *processorImpl) PostXML(apiRequest *JenkinsAPIRequest, receiver result.Result) error {
	var err error
	var httpRequest *http.Request

	httpRequest, err = processor.fb.newHTTPRequest(apiRequest)
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
func (processor *processorImpl) setCrumbs(httpRequest *http.Request) error {
	var err error
	var crumbRequest *http.Request

	crumbRequest, err = processor.fb.newCrumbRequest()
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
func (processor *processorImpl) processRequest(httpRequest *http.Request, receiver result.Result) error {
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
func NewProcessor(baseURL string, username string, password string, debug bool) (Processor, error) {

	var (
		err       error
		cookieJar *cookiejar.Jar
		transport *http.Transport
		client    *http.Client
		fb        *fabric
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
	fb = &fabric{baseURL, username, password}

	return &processorImpl{client, fb, debug}, nil
}
