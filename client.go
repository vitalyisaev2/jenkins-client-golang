package jenkins

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/vitalyisaev2/jenkins-client-golang/request"
)

// Jenkins is an access point to Jenkins API
type Client interface {
	// RootInfo returns basic information about the node that you've connected to
	RootInfo(ctx context.Context) (*Root, error)
	// JobCreate creates new job with  given name and xml configuration dumped into string
	JobCreate(ctx context.Context, name, config string) (*Job, error)
	// JobGet requests common job information for a given job name
	JobGet(ctx context.Context, name string, depth int) (*Job, error)
	// JobDelete deletes the requested job
	JobDelete(ctx context.Context, name string) error
	// JobExists checks wether job with a given name exists or not
	JobExists(ctx context.Context, name string) (bool, error)
	// JobInQueue checks whether job with a given name is in queue at the moment
	JobInQueue(ctx context.Context, name string) (bool, error)
	// JobIsBuilding checks whether job with a given name is building at the moment
	JobIsBuilding(ctx context.Context, name string) (bool, error)
	// BuildInvoke invokes simple (non-paramethrized) build of a given job
	BuildInvoke(ctx context.Context, name string) (*BuildInvoked, error)
	// BuildGetByNumber returns information about particular jenkins build
	BuildGetByNumber(ctx context.Context, name string, id int) (*Build, error)
	// BuildGetByNumber returns information about particular jenkins build by given queue id
	BuildGetByQueueID(ctx context.Context, name string, id int) (*Build, error)
}

type defaultClient struct {
	processor request.Processor
}

func (c *defaultClient) RootInfo(ctx context.Context) (*Root, error) {
	var receiver Root
	apiRequest := &request.JenkinsAPIRequest{
		Method:     "GET",
		Route:      "",
		Format:     request.JenkinsAPIFormatJSON,
		DumpMethod: request.ResponseDumpDefaultJSON,
	}
	err := c.processor.GetJSON(apiRequest, &receiver)
	return &receiver, err
}

func (c *defaultClient) JobCreate(ctx context.Context, name, config string) (*Job, error) {
	params := map[string]string{
		"name": name,
	}

	apiRequest := &request.JenkinsAPIRequest{
		Method:      "POST",
		Route:       "/createItem",
		Format:      request.JenkinsAPIFormatJSON,
		Body:        strings.NewReader(config),
		QueryParams: params,
		DumpMethod:  request.ResponseDumpDefaultJSON,
	}

	if err := c.processor.PostXML(apiRequest, nil); err != nil {
		return nil, err
	}
	return c.JobGet(ctx, name, 0)
}

func (c *defaultClient) JobGet(ctx context.Context, name string, depth int) (*Job, error) {
	var (
		receiver Job
		params   map[string]string
	)

	if depth > 2 {
		return nil, fmt.Errorf("Invalid depth")
	}
	if depth != 0 {
		params["depth"] = strconv.FormatUint(uint64(depth), 10)
	}

	apiRequest := &request.JenkinsAPIRequest{
		Method:      "GET",
		Route:       fmt.Sprintf("/job/%s", name),
		Format:      request.JenkinsAPIFormatJSON,
		QueryParams: params,
		DumpMethod:  request.ResponseDumpDefaultJSON,
	}
	err := c.processor.GetJSON(apiRequest, &receiver)
	return &receiver, err
}

func (c *defaultClient) JobDelete(ctx context.Context, name string) error {
	apiRequest := request.JenkinsAPIRequest{
		Method:     "POST",
		Route:      fmt.Sprintf("/job/%s/doDelete", name),
		Format:     request.JenkinsAPIFormatJSON,
		DumpMethod: request.ResponseDumpNone,
	}
	return c.processor.Post(&apiRequest, nil)
}

func (c *defaultClient) JobExists(ctx context.Context, name string) (bool, error) {
	info, err := c.RootInfo(ctx)
	if err != nil {
		return false, err
	}
	for _, job := range info.Jobs {
		if job.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (c *defaultClient) JobInQueue(ctx context.Context, name string) (bool, error) {
	job, err := c.JobGet(ctx, name, 0)
	if err != nil {
		return false, err
	}
	return job.InQueue, nil
}

func (c *defaultClient) JobIsBuilding(ctx context.Context, name string) (bool, error) {
	job, err := c.JobGet(ctx, name, 0)
	if err != nil {
		return false, err
	}
	return job.LastBuild.Building, nil
}

func (c *defaultClient) BuildInvoke(ctx context.Context, name string) (*BuildInvoked, error) {
	var receiver url.URL
	apiRequest := &request.JenkinsAPIRequest{
		Method:     "POST",
		Route:      fmt.Sprintf("/job/%s/build", name),
		Format:     request.JenkinsAPIFormatJSON,
		DumpMethod: request.ResponseDumpHeaderLocation,
	}
	if err := c.processor.Post(apiRequest, &receiver); err != nil {
		return nil, err
	}
	return NewBuildInvokedFromURL(&receiver)
}

func (c *defaultClient) BuildGetByNumber(ctx context.Context, name string, buildID int) (*Build, error) {
	var receiver Build
	apiRequest := &request.JenkinsAPIRequest{
		Method:      "GET",
		Route:       fmt.Sprintf("/job/%s/%d", name, buildID),
		Format:      request.JenkinsAPIFormatJSON,
		Body:        nil,
		QueryParams: nil,
		DumpMethod:  request.ResponseDumpDefaultJSON,
	}
	err := c.processor.GetJSON(apiRequest, &receiver)
	return &receiver, err
}

// auxiliary data types for BuildGetByQueueID request
type build struct {
	BuildID string `json:"id"`
	QueueID int    `json:"queueId"`
	URL     string `json:"url"`
}
type buildList struct {
	Builds []*build `json:"builds"`
}

func (c *defaultClient) BuildGetByQueueID(ctx context.Context, name string, queueID int) (*Build, error) {
	// 1. Request list of brief build descriptions of a particular job
	var (
		receiver buildList
		params   map[string]string
	)
	params["tree"] = "builds[id,queueId,url]"

	apiRequest := &request.JenkinsAPIRequest{
		Method:      "GET",
		Route:       fmt.Sprintf("/job/%s", name),
		Format:      request.JenkinsAPIFormatJSON,
		Body:        nil,
		QueryParams: params,
		DumpMethod:  request.ResponseDumpDefaultJSON,
	}
	if err := c.processor.GetJSON(apiRequest, &receiver); err != nil {
		return nil, err
	}

	// 2. Search for a job with a particular queueID
	var (
		err     error
		buildID int
	)
	for _, item := range receiver.Builds {
		if queueID == item.QueueID {
			if buildID, err = strconv.Atoi(item.BuildID); err != nil {
				return nil, err
			}
			break
		}
	}
	if buildID == 0 {
		return nil, fmt.Errorf("Build for a job %s with a queueID %d was not found", name, queueID)
	}

	// 3. Get build
	return c.BuildGetByNumber(ctx, name, buildID)
}

// PluginInstall performs installation of latest version of the plugin to Jenkins server;
// paricular version cannot be specified, see https://issues.jenkins-ci.org/browse/JENKINS-32793
func (c *defaultClient) PluginInstall(ctx context.Context, name string) error {
	body := strings.NewReader(
		fmt.Sprintf(`<jenkins><install plugin="%s@current"></jenkins>`, name),
	)
	apiRequest := request.JenkinsAPIRequest{
		Method:      "POST",
		Route:       "/pluginManager/uploadPlugin",
		Format:      request.JenkinsAPIFormatJSON,
		Body:        body,
		QueryParams: nil,
		DumpMethod:  request.ResponseDumpNone,
	}
	return c.processor.Post(&apiRequest, nil)
}

// NewJenkins initialises an entrypoint for Jenkins API
func NewClient(baseURL string, username string, password string, debug bool) (Client, error) {

	processor, err := request.NewProcessor(baseURL, username, password, debug)
	if err != nil {
		return nil, err
	}

	return &defaultClient{processor: processor}, nil
}
