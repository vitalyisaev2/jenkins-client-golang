package response

// Queue
type Queue struct {
	Items []QueueItem `json:"items"`
}

// QueueItem
type QueueItem struct {
	Actions                    []struct{} `json:"actions"`
	Blocked                    bool       `json:"blocked"`
	Buildable                  bool       `json:"buildable"`
	BuildableStartMilliseconds uint       `json:"buildableStartMilliseconds"`
	ID                         uint       `json:"id"`
	InQueueSince               uint       `json:"inQueueSince"`
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
