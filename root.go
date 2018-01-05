package jenkins

// Root represents common information about Jenkins node
type Root struct {
	AssignedLabels  []struct{}  `json:"assignedLabels"`
	Mode            string      `json:"mode"`
	NodeDescription string      `json:"nodeDescription"`
	NodeName        string      `json:"nodeName"`
	NumExecutors    uint8       `json:"numExecutors"`
	Jobs            []JobBrief  `json:"jobs"`
	OverallLoad     struct{}    `json:"overallLoad"`
	PrimaryView     PrimaryView `json:"primaryView"`
	QuietingDown    bool        `json:"quietingDown"`
	SlaveAgentPort  uint32      `json:"slaveAgentPort"`
	UnlabeledLoad   struct{}    `json:"unlabeledLoad"`
	UseCrumbs       bool        `json:"useCrumbs"`
	UseSecurity     bool        `json:"useSecurity"`
	Views           []View      `json:"views"`
}
