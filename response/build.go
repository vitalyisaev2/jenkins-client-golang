package response

// Build ???
type Build struct {
	Actions   []BuildAction
	Artifacts []struct {
		DisplayPath  string `json:"displayPath"`
		FileName     string `json:"fileName"`
		RelativePath string `json:"relativePath"`
	} `json:"artifacts"`
	Building  bool   `json:"building"`
	BuiltOn   string `json:"builtOn"`
	ChangeSet struct {
		Items []struct {
			AffectedPaths []string `json:"affectedPaths"`
			Author        struct {
				AbsoluteURL string `json:"absoluteUrl"`
				FullName    string `json:"fullName"`
			} `json:"author"`
			Comment  string `json:"comment"`
			CommitID string `json:"commitId"`
			Date     string `json:"date"`
			ID       string `json:"id"`
			Msg      string `json:"msg"`
			Paths    []struct {
				EditType string `json:"editType"`
				File     string `json:"file"`
			} `json:"paths"`
			Timestamp uint `json:"timestamp"`
		} `json:"items"`
		Kind      string `json:"kind"`
		Revisions []struct {
			Module   string
			Revision int
		} `json:"revision"`
	} `json:"changeSet"`
	Culprits          []BuildCuilprit `json:"culprits"`
	Description       interface{}     `json:"description"`
	Duration          uint            `json:"duration"`
	EstimatedDuration uint            `json:"estimatedDuration"`
	Executor          interface{}     `json:"executor"`
	FullDisplayName   string          `json:"fullDisplayName"`
	ID                string          `json:"id"`
	KeepLog           bool            `json:"keepLog"`
	Number            uint            `json:"number"`
	Result            string          `json:"result"`
	Timestamp         uint            `json:"timestamp"`
	URL               string          `json:"url"`
	MavenArtifacts    interface{}     `json:"mavenArtifacts"`
	MavenVersionUsed  string          `json:"mavenVersionUsed"`
	Runs              []struct {
		Number uint
		URL    string
	} `json:"runs"`
}

// BuildParameter ???
type BuildParameter struct {
	Name  string
	Value string
}

// BuildBranch ???
type BuildBranch struct {
	SHA1 string
	Name string
}

// BuildRevision ???
type BuildRevision struct {
	SHA1   string        `json:"SHA1"`
	Branch []BuildBranch `json:"branch"`
}

// Builds ???
type Builds struct {
	BuildNumber uint          `json:"buildNumber"`
	BuildResult interface{}   `json:"buildResult"`
	Marked      BuildRevision `json:"marked"`
	Revision    BuildRevision `json:"revision"`
}

// BuildCuilprit ???
type BuildCuilprit struct {
	AbsoluteURL string
	FullName    string
}

// BuildAction ???
type BuildAction struct {
	Parameters              []BuildParameter         `json:"parameters"`
	Causes                  []map[string]interface{} `json:"causes"`
	BuildsByBranchName      map[string]Builds        `json:"buildsByBranchName"`
	LastBuiltRevision       BuildRevision            `json:"lastBuiltRevision"`
	RemoteURLs              []string                 `json:"remoteUrls"`
	ScmName                 string                   `json:"scmName"`
	MercurialNodeName       string                   `json:"mercurialNodeName"`
	MercurialRevisionNumber string                   `json:"mercurialRevisionNumber"`
	Subdir                  interface{}              `json:"subdir"`
	TotalCount              uint
	URLName                 string
}
