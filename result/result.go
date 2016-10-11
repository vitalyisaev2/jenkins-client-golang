package result

import (
	"github.com/vitalyisaev2/jenkins-client-golang/response"
)

// Result is a common interface for API responses
// TODO: is it necessary at all
type Result interface{}

// Root represents common information about Jenkins node
// route: /api/json
type Root struct {
	Response *response.Root
	Error    error
}

// Job represents the common job item existing on Jenkins
// route: /job/jobName
type Job struct {
	Response *response.Job
	Error    error
}

// Build represents information about job build
// route: /job/jobName
type Build struct {
	Response *response.Build
	Error    error
}
