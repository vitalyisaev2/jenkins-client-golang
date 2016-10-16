package result

import (
	"github.com/vitalyisaev2/jenkins-client-golang/response"
)

// Result is a common interface for API responses
// TODO: is it necessary at all
type Result interface{}

// Root represents common information about Jenkins node
// method: GET
// route: /
type Root struct {
	Response *response.Root
	Error    error
}

// Job represents the common job item existing on Jenkins
// method: GET
// route: /job/{.jobName}
type Job struct {
	Response *response.Job
	Error    error
}

// Build represents information about job build
// method: GET
// route: /job/{.jobName}/{.buildNumber}
type Build struct {
	Response *response.Build
	Error    error
}

// BuildInvoked contains the queue position of invoked build job
// method: POST
// route: /job/{.jobName}/Build
type BuildInvoked struct {
	Response *response.BuildInvoked
	Error    error
}

// Bool response will
type Bool struct {
	Response bool
	Error    error
}
