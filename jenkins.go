package jenkins

// Jenkins is an access point to Jenkins API
type Jenkins interface {
	RootInfo() <-chan *ResultRoot
}

type jenkinsImpl struct {
	processor *requestProcessor
}

// RootInfo returns basic information about the node that you've connected to
func (j *jenkinsImpl) RootInfo() <-chan *ResultRoot {
	var responseReceiver responseRoot
	ch := make(chan *ResultRoot)
	go func() {
		defer close(ch)
		apiRequest := jenkinsAPIRequest{
			method:      "GET",
			route:       "/",
			format:      jenkinsAPIFormatJSON,
			body:        nil,
			queryParams: nil,
		}
		err := j.processor.getJSON(&apiRequest, &responseReceiver)
		if err != nil {
			ch <- &ResultRoot{nil, err}
		} else {
			ch <- &ResultRoot{&responseReceiver, nil}
		}
	}()

	return ch
}

// JobCreate tries to create new job for a given name using the dumped
//func (j *jenkinsImpl) JobCreate(jobName string, jobConfig []byte) <-chan *ResultCreateJob {

//}

// NewJenkins initialises an entrypoint for Jenkins API
func NewJenkins(baseURL string, username string, password string) (Jenkins, error) {

	processor, err := newRequestProcessor(baseURL, username, password)
	if err != nil {
		return nil, err
	}

	return &jenkinsImpl{processor}, nil
}
