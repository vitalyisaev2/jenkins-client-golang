package jenkins

import (
	"bytes"
	"fmt"

	"github.com/vitalyisaev2/jenkins-client-golang/request"
	"github.com/vitalyisaev2/jenkins-client-golang/response"
	"github.com/vitalyisaev2/jenkins-client-golang/result"
)

// Jenkins is an access point to Jenkins API
type Jenkins interface {
	RootInfo() <-chan *result.Root
	JobCreate(string, []byte) <-chan error
	JobGet(string) <-chan *result.Job
	JobDelete(string) <-chan error
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

// JobCreate creates new job for given name and xml configuration dumped into slice of bytes
func (j *jenkinsImpl) JobCreate(jobName string, jobConfig []byte) <-chan error {
	params := make(map[string]string)
	params["name"] = jobName
	body := bytes.NewBuffer(jobConfig)
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

// JobGet requests common job information for a given job name
func (j *jenkinsImpl) JobGet(jobName string) <-chan *result.Job {
	var receiver response.Job
	ch := make(chan *result.Job)

	go func() {
		defer close(ch)
		apiRequest := request.JenkinsAPIRequest{
			Method:      "GET",
			Route:       fmt.Sprintf("/job/%s", jobName),
			Format:      request.JenkinsAPIFormatJSON,
			Body:        nil,
			QueryParams: nil,
		}
		err := j.processor.GetJSON(&apiRequest, &receiver)
		if err != nil {
			ch <- &result.Job{nil, err}
		} else {
			ch <- &result.Job{&receiver, nil}
		}
	}()
	return ch
}

// JobDelete deletes the requested job
func (j *jenkinsImpl) JobDelete(jobName string) <-chan error {
	ch := make(chan error)

	go func() {
		defer close(ch)
		apiRequest := request.JenkinsAPIRequest{
			Method:      "POST",
			Route:       fmt.Sprintf("/job/%s/doDelete", jobName),
			Format:      request.JenkinsAPIFormatJSON,
			Body:        nil,
			QueryParams: nil,
		}
		ch <- j.processor.Post(&apiRequest, nil)
	}()
	return ch
}

//

// NewJenkins initialises an entrypoint for Jenkins API
func NewJenkins(baseURL string, username string, password string, debug bool) (Jenkins, error) {

	processor, err := request.NewProcessor(baseURL, username, password, debug)
	if err != nil {
		return nil, err
	}

	return &jenkinsImpl{processor}, nil
}
