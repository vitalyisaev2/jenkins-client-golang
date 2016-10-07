package jenkins

// Jenkins is an access point to Jenkins API
type Jenkins interface {
	RootInfo() <-chan *ResultRoot
}

type jenkinsImpl struct {
	processor *requestProcessor
}

func (j *jenkinsImpl) RootInfo() <-chan *ResultRoot {
	var responseReceiver responseRoot
	ch := make(chan *ResultRoot)
	go func() {
		defer close(ch)
		err := j.processor.getJSON("/", &responseReceiver)
		if err != nil {
			ch <- &ResultRoot{nil, err}
		} else {
			ch <- &ResultRoot{&responseReceiver, nil}
		}
	}()

	return ch
}

// NewJenkins initialises an entrypoint for Jenkins API
func NewJenkins(baseURL string, username string, password string) (Jenkins, error) {

	processor, err := newRequestProcessor(baseURL, username, password)
	if err != nil {
		return nil, err
	}

	return &jenkinsImpl{processor}, nil
}
