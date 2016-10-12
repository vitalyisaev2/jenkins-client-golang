package jenkins

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"

	"github.com/vitalyisaev2/jenkins-client-golang/request"
	"github.com/vitalyisaev2/jenkins-client-golang/response"
	"github.com/vitalyisaev2/jenkins-client-golang/result"
)

// Jenkins is an access point to Jenkins API
type Jenkins interface {
	RootInfo() <-chan *result.Root
	JobCreate(string, []byte) <-chan *result.Job
	JobGet(string) <-chan *result.Job
	JobDelete(string) <-chan error
	BuildInvoke(string) <-chan *result.QueueItem
	BuildGetByNumber(string, uint) <-chan *result.Build
	BuildGetByQueueID(string, uint) <-chan *result.Build
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
			DumpMethod:  request.ResponseDumpDefaultJSON,
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
func (j *jenkinsImpl) JobCreate(jobName string, jobConfig []byte) <-chan *result.Job {
	var err error
	params := make(map[string]string)
	params["name"] = jobName
	body := bytes.NewBuffer(jobConfig)
	ch := make(chan *result.Job)

	go func() {
		defer close(ch)
		apiRequest := request.JenkinsAPIRequest{
			Method:      "POST",
			Route:       "/createItem",
			Format:      request.JenkinsAPIFormatJSON,
			Body:        body,
			QueryParams: params,
			DumpMethod:  request.ResponseDumpDefaultJSON,
		}
		err = j.processor.PostXML(&apiRequest, nil)
		switch err {
		case nil:
			res := <-j.JobGet(jobName)
			ch <- res
		default:
			ch <- &result.Job{nil, err}
		}
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
			DumpMethod:  request.ResponseDumpDefaultJSON,
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
			DumpMethod:  request.ResponseDumpNone,
		}
		ch <- j.processor.Post(&apiRequest, nil)
	}()
	return ch
}

func (j *jenkinsImpl) BuildInvoke(jobName string) <-chan *result.QueueItem {
	var receiver url.URL
	ch := make(chan *result.QueueItem)

	go func() {
		defer close(ch)
		apiRequest := request.JenkinsAPIRequest{
			Method:      "POST",
			Route:       fmt.Sprintf("/job/%s/build", jobName),
			Format:      request.JenkinsAPIFormatJSON,
			Body:        nil,
			QueryParams: nil,
			DumpMethod:  request.ResponseDumpHeaderLocation,
		}
		err := j.processor.Post(&apiRequest, &receiver)
		if err != nil {
			ch <- &result.QueueItem{nil, err}
		} else {
			if resp, err := response.NewQueueItemFromURL(&receiver); err != nil {
				ch <- &result.QueueItem{nil, err}
			} else {
				ch <- &result.QueueItem{resp, nil}
			}
		}
	}()
	return ch
}

func (j *jenkinsImpl) BuildGetByNumber(jobName string, buildNumber uint) <-chan *result.Build {
	var receiver response.Build

	ch := make(chan *result.Build)
	go func() {
		defer close(ch)
		apiRequest := request.JenkinsAPIRequest{
			Method:      "GET",
			Route:       fmt.Sprintf("/job/%s/%d", jobName, buildNumber),
			Format:      request.JenkinsAPIFormatJSON,
			Body:        nil,
			QueryParams: nil,
			DumpMethod:  request.ResponseDumpDefaultJSON,
		}
		err := j.processor.GetJSON(&apiRequest, &receiver)
		if err != nil {
			ch <- &result.Build{nil, err}
		} else {
			ch <- &result.Build{&receiver, nil}
		}
	}()
	return ch
}

func (j *jenkinsImpl) BuildGetByQueueID(jobName string, queueID uint) <-chan *result.Build {
	var err error

	ch := make(chan *result.Build)
	go func() {
		defer close(ch)

		// 1. Request list of brief build descriptions for a particular job
		type buildListItem struct {
			BuildID string `json:"id"`
			QueueID uint   `json:"queueId"`
			URL     string `json:"url"`
		}
		type buildList struct {
			Builds []buildListItem `json:"builds"`
		}
		var buildListReceiver buildList
		buildListParams := make(map[string]string)
		buildListParams["tree"] = "builds[id,queueId,url]"

		buildListRequest := request.JenkinsAPIRequest{
			Method:      "GET",
			Route:       fmt.Sprintf("/job/%s", jobName),
			Format:      request.JenkinsAPIFormatJSON,
			Body:        nil,
			QueryParams: buildListParams,
			DumpMethod:  request.ResponseDumpDefaultJSON,
		}
		err = j.processor.GetJSON(&buildListRequest, &buildListReceiver)
		if err != nil {
			ch <- &result.Build{nil, err}
			return
		}

		// 2. Search for a job with a particular queueID
		var targetBuildNumber uint
		for _, item := range buildListReceiver.Builds {
			if queueID == item.QueueID {
				var u64 uint64
				if u64, err = strconv.ParseUint(item.BuildID, 10, 0); err != nil {
					ch <- &result.Build{nil, err}
					return
				}
				targetBuildNumber = uint(u64)
				break
			}
		}
		if targetBuildNumber == 0 {
			ch <- &result.Build{
				nil,
				fmt.Errorf("Build for a job %s with a queueID %d was not found", jobName, queueID),
			}
			return
		}

		// 3. Get job with a particular build
		resp := <-j.BuildGetByNumber(jobName, targetBuildNumber)
		ch <- resp
	}()
	return ch
}

// NewJenkins initialises an entrypoint for Jenkins API
func NewJenkins(baseURL string, username string, password string, debug bool) (Jenkins, error) {

	processor, err := request.NewProcessor(baseURL, username, password, debug)
	if err != nil {
		return nil, err
	}

	return &jenkinsImpl{processor}, nil
}
