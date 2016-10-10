package jenkins

import (
	"bytes"
	"io"

	"github.com/vitalyisaev2/jenkins-client-golang/request"
	"github.com/vitalyisaev2/jenkins-client-golang/response"
	"github.com/vitalyisaev2/jenkins-client-golang/result"
)

// Jenkins is an access point to Jenkins API
type Jenkins interface {
	RootInfo() <-chan *result.Root
	JobCreate(string, []byte) <-chan error
}

type jenkinsImpl struct {
	processor request.Processor
}

// RootInfo returns basic information about the node that you've connected to
func (j *jenkinsImpl) RootInfo() <-chan *result.Root {
	var receiver response.Root
	ch := make(chan *result.Root)
	go func() {
		defer close(ch)
		apiRequest := request.JenkinsAPIRequest{
			Method:      "GET",
			Route:       "",
			Format:      request.JenkinsAPIFormatJSON,
			Body:        nil,
			QueryParams: nil,
		}
		err := j.processor.GetJSON(&apiRequest, &receiver)
		if err != nil {
			ch <- &result.Root{nil, err}
		} else {
			ch <- &result.Root{&receiver, nil}
		}
	}()

	return ch
}

// JobCreate creates new job for a given name using the xml configuration dumped into slice of bytes
func (j *jenkinsImpl) JobCreate(jobName string, jobConfig []byte) <-chan error {
	var body io.Reader

	params := make(map[string]string)
	params["name"] = jobName

	body = bytes.NewBuffer(jobConfig)

	ch := make(chan error)
	go func() {
		defer close(ch)
		apiRequest := request.JenkinsAPIRequest{
			Method:      "POST",
			Route:       "/createItem",
			Format:      request.JenkinsAPIFormatJSON,
			Body:        body,
			QueryParams: params,
		}
		ch <- j.processor.PostXML(&apiRequest, nil)
	}()
	return ch
}

// JobGet get

// NewJenkins initialises an entrypoint for Jenkins API
func NewJenkins(baseURL string, username string, password string, debug bool) (Jenkins, error) {

	processor, err := request.NewProcessor(baseURL, username, password, debug)
	if err != nil {
		return nil, err
	}

	return &jenkinsImpl{processor}, nil
}
