package jenkins

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
)

// Jenkins is an access point to Jenkins API
type Jenkins interface {
	RootInfo() <-chan *Result
}

type jenkinsImpl struct {
	client *http.Client
	rb     *requestBuilder
	//username string
	//password string
}

func (j *jenkinsImpl) getJSON(route string, responseReceiver APIResponse) error {
	var err error
	var req *http.Request
	var resp *http.Response

	req, err = j.rb.newJSONRequest("GET", route, nil)
	if err != nil {
		return err
	}

	resp, err = j.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		location, _ := resp.Location()
		return fmt.Errorf("%v: %s", location, resp.Status)
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
func NewJenkins(baseURL string, username string, password string) (Jenkins, error) {
	var (
		err       error
		cookieJar *cookiejar.Jar
		transport *http.Transport
		client    *http.Client
		rb        *requestBuilder
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

	client = &http.Client{
		Transport: transport,
		Jar:       cookieJar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			req.SetBasicAuth(username, password)
			return nil
		},
	}

	rb = &requestBuilder{baseURL, username, password}

	jenkins = &jenkinsImpl{client, rb}

	return jenkins, nil
}

func redirectPolicyFunc(req *http.Request, via []*http.Request) error {

	return nil
}
