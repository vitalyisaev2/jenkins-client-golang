package jenkins

// Queue contains list of invoked builds waiting for available executor
type Queue struct {
	Items []QueueItem `json:"items"`
}

// QueueItem represents invoked build
type QueueItem struct {
	Actions                    []struct{} `json:"actions"`
	Blocked                    bool       `json:"blocked"`
	Buildable                  bool       `json:"buildable"`
	BuildableStartMilliseconds int        `json:"buildableStartMilliseconds"`
	ID                         int        `json:"id"`
	InQueueSince               int        `json:"inQueueSince"`
	Params                     string     `json:"params"`
	Pending                    bool       `json:"pending"`
	Stuck                      bool       `json:"stuck"`
	Task                       struct {
		Color string `json:"color"`
		Name  string `json:"name"`
		URL   string `json:"url"`
	} `json:"task"`
	URL string `json:"url"`
	Why string `json:"why"`
}
