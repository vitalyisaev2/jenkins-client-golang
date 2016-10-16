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
	// RootInfo returns basic information about the node that you've connected to
	RootInfo() <-chan *result.Root
	// JobCreate creates new job for given name and xml configuration dumped into slice of bytes
	JobCreate(string, []byte) <-chan *result.Job
	// JobGet requests common job information for a given job name
	JobGet(string, uint) <-chan *result.Job
	// JobDelete deletes the requested job
	JobDelete(string) <-chan error
	// JobExists checks wether job with a given name exists or not
	JobExists(string) <-chan *result.Bool
	// JobInQueue checks whether job with a given name is in queue at the moment
	JobInQueue(string) <-chan *result.Bool
	// JobIsBuilding checks whether job with a given name is building at the moment
	JobIsBuilding(string) <-chan *result.Bool
	// BuildInvoke invokes simple (non-paramethrized) build of a given job
	BuildInvoke(string) <-chan *result.BuildInvoked
	// BuildGetByNumber returns information about particular jenkins build
	BuildGetByNumber(string, uint) <-chan *result.Build
	// BuildGetByNumber returns information about particular jenkins build by given queue id
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
			ch <- &result.Root{Response: nil, Error: err}
		} else {
			ch <- &result.Root{Response: &receiver, Error: nil}
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
			res := <-j.JobGet(jobName, 0)
			ch <- res
		default:
			ch <- &result.Job{Response: nil, Error: err}
		}
	}()
	return ch
}

// JobGet requests common job information for a given job name
func (j *jenkinsImpl) JobGet(jobName string, depth uint) <-chan *result.Job {
	var receiver response.Job
	ch := make(chan *result.Job)

	go func() {
		defer close(ch)
		params := make(map[string]string)

		if depth > 2 {
			ch <- &result.Job{Response: nil, Error: fmt.Errorf("Bad depth")}
		} else if depth != 0 {
			params["depth"] = strconv.FormatUint(uint64(depth), 10)
		}

		apiRequest := request.JenkinsAPIRequest{
			Method:      "GET",
			Route:       fmt.Sprintf("/job/%s", jobName),
			Format:      request.JenkinsAPIFormatJSON,
			Body:        nil,
			QueryParams: params,
			DumpMethod:  request.ResponseDumpDefaultJSON,
		}
		err := j.processor.GetJSON(&apiRequest, &receiver)
		if err != nil {
			ch <- &result.Job{Response: nil, Error: err}
		} else {
			ch <- &result.Job{Response: &receiver, Error: nil}
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

// JobExists checks wether job with a given name exists or not
func (j *jenkinsImpl) JobExists(jobName string) <-chan *result.Bool {
	ch := make(chan *result.Bool)
	go func() {
		defer close(ch)
		info := <-j.RootInfo()
		if info.Error != nil {
			ch <- &result.Bool{Error: info.Error}
			return
		}
		for _, job := range info.Response.Jobs {
			if job.Name == jobName {
				ch <- &result.Bool{Response: true, Error: nil}
				return
			}
		}
		ch <- &result.Bool{Response: false, Error: nil}
	}()
	return ch
}

// JobExists checks wether job with a given name exists or not
func (j *jenkinsImpl) JobInQueue(jobName string) <-chan *result.Bool {
	ch := make(chan *result.Bool)
	go func() {
		defer close(ch)
		jobGet := <-j.JobGet(jobName, 0)
		if jobGet.Error != nil {
			ch <- &result.Bool{Error: jobGet.Error}
			return
		}
		ch <- &result.Bool{Response: jobGet.Response.InQueue, Error: jobGet.Error}
	}()
	return ch
}

// JobIsBuilding checks whether job with a given name is building at the moment
func (j *jenkinsImpl) JobIsBuilding(jobName string) <-chan *result.Bool {
	ch := make(chan *result.Bool)
	go func() {
		defer close(ch)
		jobGet := <-j.JobGet(jobName, 1)
		if jobGet.Error != nil {
			ch <- &result.Bool{Error: jobGet.Error}
			return
		}
		//fmt.Println("JobIsBuilding: ", jobGet.Response.LastBuild.Building)
		//ch <- jobGet.Response.LastBuild.Building
		ch <- &result.Bool{Response: jobGet.Response.LastBuild.Building, Error: nil}
	}()
	return ch
}

// BuildInvoke invokes simple (non-paramethrized) build of a given job
func (j *jenkinsImpl) BuildInvoke(jobName string) <-chan *result.BuildInvoked {
	var receiver url.URL
	ch := make(chan *result.BuildInvoked)

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
			ch <- &result.BuildInvoked{Response: nil, Error: err}
		} else {
			if resp, err := response.NewBuildInvokedFromURL(&receiver); err != nil {
				ch <- &result.BuildInvoked{Response: nil, Error: err}
			} else {
				ch <- &result.BuildInvoked{Response: resp, Error: nil}
			}
		}
	}()
	return ch
}

// BuildGetByNumber returns information about particular jenkins build
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
			ch <- &result.Build{Response: nil, Error: err}
		} else {
			ch <- &result.Build{Response: &receiver, Error: nil}
		}
	}()
	return ch
}

// BuildGetByQueueID returns information about particular jenkins build by given queue id
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
			ch <- &result.Build{Response: nil, Error: err}
			return
		}

		// 2. Search for a job with a particular queueID
		var targetBuildNumber uint
		for _, item := range buildListReceiver.Builds {
			if queueID == item.QueueID {
				var u64 uint64
				if u64, err = strconv.ParseUint(item.BuildID, 10, 0); err != nil {
					ch <- &result.Build{Response: nil, Error: err}
					return
				}
				targetBuildNumber = uint(u64)
				break
			}
		}
		if targetBuildNumber == 0 {
			ch <- &result.Build{
				Response: nil,
				Error:    fmt.Errorf("Build for a job %s with a queueID %d was not found", jobName, queueID),
			}
			return
		}

		// 3. Get job with a particular build
		resp := <-j.BuildGetByNumber(jobName, targetBuildNumber)
		ch <- resp
	}()
	return ch
}

//PluginInstall performs installation of latest version of the plugin in Jenkins.
//paricular version cannot be specified, see https://issues.jenkins-ci.org/browse/JENKINS-32793
func (j *jenkinsImpl) PluginInstall(pluginName string) <-chan error {
	bodyStr := fmt.Sprintf(`<jenkins><install plugin="%s@current"></jenkins>`, pluginName)
	body := bytes.NewBufferString(bodyStr)
	ch := make(chan error)

	go func() {
		defer close(ch)
		apiRequest := request.JenkinsAPIRequest{
			Method:      "POST",
			Route:       "/pluginManager/uploadPlugin",
			Format:      request.JenkinsAPIFormatJSON,
			Body:        body,
			QueryParams: nil,
			DumpMethod:  request.ResponseDumpNone,
		}
		ch <- j.processor.Post(&apiRequest, nil)
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
