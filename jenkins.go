package jenkins

import (
	"bytes"
	"io"
)

// Jenkins is an access point to Jenkins API
type Jenkins interface {
	RootInfo() <-chan *ResultRoot
	JobCreate(string, []byte) <-chan error
}

type jenkinsImpl struct {
	processor *requestProcessor
}

// RootInfo returns basic information about the node that you've connected to
func (j *jenkinsImpl) RootInfo() <-chan *ResultRoot {
	var receiver ResponseRoot
	ch := make(chan *ResultRoot)
	go func() {
		defer close(ch)
		apiRequest := jenkinsAPIRequest{
			method:      "GET",
			route:       "",
			format:      jenkinsAPIFormatJSON,
			body:        nil,
			queryParams: nil,
		}
		err := j.processor.getJSON(&apiRequest, &receiver)
		if err != nil {
			ch <- &ResultRoot{nil, err}
		} else {
			ch <- &ResultRoot{&receiver, nil}
		}
	}()

	return ch
}

// JobCreate to create new job for a given name using the xml configuration dumped into byte slice
func (j *jenkinsImpl) JobCreate(jobName string, jobConfig []byte) <-chan error {
	var body io.Reader

	params := make(map[string]string)
	params["name"] = jobName

	body = bytes.NewBuffer(jobConfig)

	ch := make(chan error)
	go func() {
		defer close(ch)
		apiRequest := jenkinsAPIRequest{
			method:      "POST",
			route:       "/createItem",
			format:      jenkinsAPIFormatJSON,
			body:        body,
			queryParams: params,
		}
		ch <- j.processor.postXML(&apiRequest, nil)
	}()
	return ch
}

// NewJenkins initialises an entrypoint for Jenkins API
func NewJenkins(baseURL string, username string, password string, debug bool) (Jenkins, error) {

	processor, err := newRequestProcessor(baseURL, username, password, debug)
	if err != nil {
		return nil, err
	}

	return &jenkinsImpl{processor}, nil
}
