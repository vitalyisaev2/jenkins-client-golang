package result

import (
	"github.com/vitalyisaev2/jenkins-client-golang/response"
)

// Result is a common interface for API responses
// TODO: is it necessary at all
type Result interface{}

// ResultRoot represents common information about Jenkins node
// route: /api/json
type Root struct {
	Response *response.Root
	Error    error
}

// ResultCreateJob represents the result of job creation
// route: /createItem?name=jobName
type CreateJob struct {
	Response *response.Job
	Error    error
}
