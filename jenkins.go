package jenkins

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
)

// Jenkins is an access point to Jenkins API
type Jenkins interface {
	RootInfo() <-chan *Result
}

type jenkinsImpl struct {
	client  *http.Client
	baseurl string
	//username string
	//password string
}

func (j *jenkinsImpl) newJSONRequest(method string, route string, body io.Reader) (*http.Request, error) {
	URL := fmt.Sprintf("%s/%s", j.baseurl, route)
	req, err := http.NewRequest(method, URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/xml")
	return req, nil
}

func (j *jenkinsImpl) getJSON(route string, responseReceiver APIResponse) error {
	var err error
	var req *http.Request
	var resp *http.Response
	defer resp.Body.Close()

	req, err = j.newJSONRequest("GET", route, nil)
	if err != nil {
		return err
	}

	resp, err = j.client.Do(req)
	if err != nil {
		return err
	}

	err = json.NewDecoder(resp.Body).Decode(responseReceiver)
	return err
}

func (j *jenkinsImpl) RootInfo() <-chan *Result {
	var responseReceiver APIResponseRoot
	ch := make(chan *Result)
	go func() {
		defer close(ch)
		err := j.getJSON("/", &responseReceiver)
		if err != nil {
			ch <- &Result{nil, err}
		} else {
			ch <- &Result{responseReceiver, nil}
		}
	}()

	return ch
}

// NewJenkins initialises an entrypoint for Jenkins API
func NewJenkins(baseurl string, username string, password string) (Jenkins, error) {
	var (
		err       error
		cookieJar *cookiejar.Jar
		transport *http.Transport
		client    *http.Client
		jenkins   *jenkinsImpl
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

	client = &http.Client{Transport: transport, Jar: cookieJar}

	jenkins = &jenkinsImpl{client, baseurl}

	return jenkins, nil
}
