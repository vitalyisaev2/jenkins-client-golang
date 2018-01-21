package jenkins

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
)

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
			Timestamp int `json:"timestamp"`
		} `json:"items"`
		Kind      string `json:"kind"`
		Revisions []struct {
			Module   string
			Revision int
		} `json:"revision"`
	} `json:"changeSet"`
	Culprits          []BuildCuilprit `json:"culprits"`
	Description       interface{}     `json:"description"`
	Duration          int             `json:"duration"`
	EstimatedDuration int             `json:"estimatedDuration"`
	Executor          interface{}     `json:"executor"`
	FullDisplayName   string          `json:"fullDisplayName"`
	ID                string          `json:"id"`
	KeepLog           bool            `json:"keepLog"`
	Number            int             `json:"number"`
	Result            string          `json:"result"`
	Timestamp         int             `json:"timestamp"`
	URL               string          `json:"url"`
	MavenArtifacts    interface{}     `json:"mavenArtifacts"`
	MavenVersionUsed  string          `json:"mavenVersionUsed"`
	Runs              []struct {
		Number int
		URL    string
	} `json:"runs"`
}

// BuildParameter ???
type BuildParameter struct {
	Name  string
	Value bool
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
	BuildNumber int           `json:"buildNumber"`
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
	TotalCount              int
	URLName                 string
}

// BuildInvoked is returned as a part of a response headers when the build is invoked
type BuildInvoked struct {
	URL *url.URL
	ID  int
}

// NewBuildInvokedFromURL returns struct with URL and parsed location
func NewBuildInvokedFromURL(URL *url.URL) (*BuildInvoked, error) {
	// TODO: use precompiled regex
	pattern := regexp.MustCompile("queue/item/(?P<id>[0-9]+)/")
	if !pattern.MatchString(URL.Path) {
		return nil, fmt.Errorf("Returned URL (%v) doesn't match expected pattern", URL)
	}
	raw := pattern.FindStringSubmatch(URL.Path)[1]
	buildID, err := strconv.Atoi(raw)
	if err != nil {
		return nil, err
	}
	return &BuildInvoked{URL: URL, ID: buildID}, nil
}
