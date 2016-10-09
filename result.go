package jenkins

// Result is a common interface for API responses
// TODO: is it necessary at all
type Result interface{}

// ResultRoot represents common information about Jenkins node
// route: /api/json
type ResultRoot struct {
	Response *ResponseRoot
	Error    error
}

// ResultCreateJob represents the result of job creation
// route: /createItem?name=jobName
type ResultCreateJob struct {
	Response *ResponseJob
	Error    error
}
