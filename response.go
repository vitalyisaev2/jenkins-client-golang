package jenkins

// Response ...
type Result interface{}

// ResultRoot ...
type ResultRoot struct {
	Response *responseRoot
	Error    error
}

// apiResponseRoot represents common information about Jenkins node
// route: /api/json
type responseRoot struct {
	AssignedLabels  []struct{}  `json:"assignedLabels"`
	Mode            string      `json:"mode"`
	NodeDescription string      `json:"nodeDescription"`
	NodeName        string      `json:"nodeName"`
	NumExecutors    uint8       `json:"numExecutors"`
	Jobs            []job       `json:"jobs"`
	OverallLoad     struct{}    `json:"overallLoad"`
	PrimaryView     primaryView `json:"primaryView"`
	QuietingDown    bool        `json:"quietingDown"`
	SlaveAgentPort  uint32      `json:"slaveAgentPort"`
	UnlabeledLoad   struct{}    `json:"unlabeledLoad"`
	UseCrumbs       bool        `json:"useCrumbs"`
	UseSecurity     bool        `json:"useSecurity"`
	Views           []view      `json:"views"`
}

type job struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Color string `json:"color"`
}

type primaryView struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type view struct {
}
