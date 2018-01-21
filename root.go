package jenkins

// Root represents common information about Jenkins node
type Root struct {
	AssignedLabels  []struct{}  `json:"assignedLabels"`
	Mode            string      `json:"mode"`
	NodeDescription string      `json:"nodeDescription"`
	NodeName        string      `json:"nodeName"`
	NumExecutors    int         `json:"numExecutors"`
	Jobs            []JobBrief  `json:"jobs"`
	OverallLoad     struct{}    `json:"overallLoad"`
	PrimaryView     PrimaryView `json:"primaryView"`
	QuietingDown    bool        `json:"quietingDown"`
	SlaveAgentPort  int         `json:"slaveAgentPort"`
	UnlabeledLoad   struct{}    `json:"unlabeledLoad"`
	UseCrumbs       bool        `json:"useCrumbs"`
	UseSecurity     bool        `json:"useSecurity"`
	Views           []View      `json:"views"`
}
