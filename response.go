package jenkins

// APIResponse ...
type APIResponse interface{}

// Result ...
type Result struct {
	Response APIResponse
	Error    error
}

// APIResponseRoot represents common information about Jenkins node
// route: /api/json
type APIResponseRoot struct {
	AssignedLabels  []struct{} `json:"assignedLabels"`
	Mode            string     `json:"mode"`
	NodeDescription string     `json:"nodeDescription"`
	NodeName        string     `json:"nodeName"`
	NumExecutors    uint8      `json:"numExecutors"`
	Jobs            []struct {
		Name  string `json:"name"`
		URL   string `json:"url"`
		Color string `json:"color"`
	} `json:"jobs"`
	OverallLoad struct{} `json:"overallLoad"`
	PrimaryView struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"primaryView"`
	QuietingDown   bool       `json:"quietingDown"`
	SlaveAgentPort uint32     `json:"slaveAgentPort"`
	UnlabeledLoad  struct{}   `json:"unlabeledLoad"`
	UseCrumbs      bool       `json:"useCrumbs"`
	UseSecurity    bool       `json:"useSecurity"`
	Views          []struct{} `json:"views"`
}
