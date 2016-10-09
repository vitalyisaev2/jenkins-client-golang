package jenkins

// ResponseRoot represents common information about Jenkins node
type ResponseRoot struct {
	AssignedLabels  []struct{}  `json:"assignedLabels"`
	Mode            string      `json:"mode"`
	NodeDescription string      `json:"nodeDescription"`
	NodeName        string      `json:"nodeName"`
	NumExecutors    uint8       `json:"numExecutors"`
	Jobs            []Job       `json:"jobs"`
	OverallLoad     struct{}    `json:"overallLoad"`
	PrimaryView     PrimaryView `json:"primaryView"`
	QuietingDown    bool        `json:"quietingDown"`
	SlaveAgentPort  uint32      `json:"slaveAgentPort"`
	UnlabeledLoad   struct{}    `json:"unlabeledLoad"`
	UseCrumbs       bool        `json:"useCrumbs"`
	UseSecurity     bool        `json:"useSecurity"`
	Views           []View      `json:"views"`
}

// ResponseJob represents the result of job API call
type ResponseJob struct {
	Actions            []interface{} `json:"actions"`
	Buildable          bool          `json:"buildable"`
	Builds             []JobBuild
	Color              string      `json:"color"`
	ConcurrentBuild    bool        `json:"concurrentBuild"`
	Description        string      `json:"description"`
	DisplayName        string      `json:"displayName"`
	DisplayNameOrNull  interface{} `json:"displayNameOrNull"`
	DownstreamProjects []Job       `json:"downstreamProjects"`
	FirstBuild         JobBuild
	HealthReport       []struct {
		Description   string `json:"description"`
		IconClassName string `json:"iconClassName"`
		IconURL       string `json:"iconUrl"`
		Score         int64  `json:"score"`
	} `json:"healthReport"`
	InQueue               bool     `json:"inQueue"`
	KeepDependencies      bool     `json:"keepDependencies"`
	LastBuild             JobBuild `json:"lastBuild"`
	LastCompletedBuild    JobBuild `json:"lastCompletedBuild"`
	LastFailedBuild       JobBuild `json:"lastFailedBuild"`
	LastStableBuild       JobBuild `json:"lastStableBuild"`
	LastSuccessfulBuild   JobBuild `json:"lastSuccessfulBuild"`
	LastUnstableBuild     JobBuild `json:"lastUnstableBuild"`
	LastUnsuccessfulBuild JobBuild `json:"lastUnsuccessfulBuild"`
	Name                  string   `json:"name"`
	NextBuildNumber       int64    `json:"nextBuildNumber"`
	Property              []struct {
		ParameterDefinitions []ParameterDefinition `json:"parameterDefinitions"`
	} `json:"property"`
	QueueItem        interface{} `json:"queueItem"`
	Scm              struct{}    `json:"scm"`
	UpstreamProjects []Job       `json:"upstreamProjects"`
	URL              string      `json:"url"`
}

// Job is a short representation of a common Jenkins item used in various API responses
type Job struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Color string `json:"color"`
}

// JobBuild is a short representation of a common Jenkins build used in various API responses
type JobBuild struct {
	Number int64
	URL    string
}

// PrimaryView ???
type PrimaryView struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// View ???
type View struct {
}

// ParameterDefinition ???
type ParameterDefinition struct {
	DefaultParameterValue struct {
		Name  string `json:"name"`
		Value bool   `json:"value"`
	} `json:"defaultParameterValue"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Type        string `json:"type"`
}
