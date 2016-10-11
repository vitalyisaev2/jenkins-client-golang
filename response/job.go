package response

// Job represents the result of job API call
type Job struct {
	Actions            []interface{} `json:"actions"`
	Buildable          bool          `json:"buildable"`
	Builds             []JobBuildBrief
	Color              string      `json:"color"`
	ConcurrentBuild    bool        `json:"concurrentBuild"`
	Description        string      `json:"description"`
	DisplayName        string      `json:"displayName"`
	DisplayNameOrNull  interface{} `json:"displayNameOrNull"`
	DownstreamProjects []JobBrief  `json:"downstreamProjects"`
	FirstBuild         JobBuildBrief
	HealthReport       []struct {
		Description   string `json:"description"`
		IconClassName string `json:"iconClassName"`
		IconURL       string `json:"iconUrl"`
		Score         uint   `json:"score"`
	} `json:"healthReport"`
	InQueue               bool          `json:"inQueue"`
	KeepDependencies      bool          `json:"keepDependencies"`
	LastBuild             JobBuildBrief `json:"lastBuild"`
	LastCompletedBuild    JobBuildBrief `json:"lastCompletedBuild"`
	LastFailedBuild       JobBuildBrief `json:"lastFailedBuild"`
	LastStableBuild       JobBuildBrief `json:"lastStableBuild"`
	LastSuccessfulBuild   JobBuildBrief `json:"lastSuccessfulBuild"`
	LastUnstableBuild     JobBuildBrief `json:"lastUnstableBuild"`
	LastUnsuccessfulBuild JobBuildBrief `json:"lastUnsuccessfulBuild"`
	Name                  string        `json:"name"`
	NextBuildNumber       uint          `json:"nextBuildNumber"`
	Property              []struct {
		ParameterDefinitions []JobParameterDefinition `json:"parameterDefinitions"`
	} `json:"property"`
	QueueItem        interface{} `json:"queueItem"`
	Scm              struct{}    `json:"scm"`
	UpstreamProjects []JobBrief  `json:"upstreamProjects"`
	URL              string      `json:"url"`
}

// JobBrief is a short representation of a common Jenkins item used in various API responses
type JobBrief struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Color string `json:"color"`
}

// JobBuildBrief is a short representation of a common Jenkins build used in various API responses
type JobBuildBrief struct {
	Number uint
	URL    string
}

// JobParameterDefinition ???
type JobParameterDefinition struct {
	DefaultParameterValue struct {
		Name  string `json:"name"`
		Value bool   `json:"value"`
	} `json:"defaultParameterValue"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Type        string `json:"type"`
}
