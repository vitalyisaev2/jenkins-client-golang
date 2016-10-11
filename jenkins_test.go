package jenkins_test

import (
	"fmt"
	"log"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalyisaev2/jenkins-client-golang"
)

const (
	baseURL           string = "http://localhost:8080"
	jenkinsAdminLogin string = "admin"
	debug             bool   = false
)

var jenkinsAdminPassword string
var api jenkins.Jenkins

// Get admin password for test purposes
func init() {
	out, err := exec.Command("docker", "exec", "jenkins", "cat", "/var/jenkins_home/secrets/initialAdminPassword").Output()
	if err != nil {
		log.Fatal(err)
	}
	jenkinsAdminPassword = string(out[:len(out)-1])
	log.Printf("jenkinsAdminPassword captured: %s\n", jenkinsAdminPassword)
}

func TestInit(t *testing.T) {
	var err error
	api, err = jenkins.NewJenkins(baseURL, jenkinsAdminLogin, jenkinsAdminPassword, debug)
	assert.NotNil(t, api)
	assert.Nil(t, err)

	result := <-api.RootInfo()
	assert.NotNil(t, result)
	assert.NotNil(t, result.Response)
	assert.Nil(t, result.Error)

	assert.NotEqual(t, 0, result.Response.NumExecutors)
}

func TestJobCreate(t *testing.T) {
	jobName := "test1"
	jobConfig := []byte(`
<project>
  <actions/>
  <description></description>
  <keepDependencies>false</keepDependencies>
  <properties/>
  <scm class="hudson.scm.NullSCM"/>
  <canRoam>true</canRoam>
  <disabled>false</disabled>
  <blockBuildWhenDownstreamBuilding>false</blockBuildWhenDownstreamBuilding>
  <blockBuildWhenUpstreamBuilding>false</blockBuildWhenUpstreamBuilding>
  <triggers/>
  <concurrentBuild>false</concurrentBuild>
  <builders>
    <hudson.tasks.Shell>
      <command>sleep 10;</command>
    </hudson.tasks.Shell>
  </builders>
  <publishers/>
  <buildWrappers/>
</project>
`)
	result := <-api.JobCreate(jobName, jobConfig)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Response)
	assert.Nil(t, result.Error)

	assert.Equal(t, jobName, result.Response.DisplayName)
	assert.Equal(t, fmt.Sprintf("%s/job/%s/", baseURL, jobName), result.Response.URL)
}

func TestJobGet(t *testing.T) {
	jobName := "test1"
	result := <-api.JobGet(jobName)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Response)
	assert.Nil(t, result.Error)

	assert.Equal(t, jobName, result.Response.DisplayName)
	assert.Equal(t, fmt.Sprintf("%s/job/%s/", baseURL, jobName), result.Response.URL)
}

//func TestJobBuild(t *testing.T) {
//}

func TestJobDelete(t *testing.T) {
	jobName := "test1"
	err := <-api.JobDelete(jobName)
	assert.Nil(t, err)
}
