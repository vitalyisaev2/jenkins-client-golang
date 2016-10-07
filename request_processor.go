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

func (processor *requestProcessor) getJSON(route string, responseReceiver Result) error {
	var err error
	var req *http.Request
	var resp *http.Response

	req, err = processor.fabric.newJSONRequest("GET", route, nil)
	if err != nil {
		return err
	}

	resp, err = processor.client.Do(req)
	if err != nil {
		return err
	}
	defer func() error {
		err = resp.Body.Close()
		return err
	}()

	if resp.StatusCode != http.StatusOK {
		location, _ := resp.Location()
		return fmt.Errorf("%v: %s", location, resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(responseReceiver)
	return err
}

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
