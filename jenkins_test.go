package jenkins_test

import (
	"fmt"
	"log"
	"os/exec"
	"testing"
	"time"

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

// Test API initialisation
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

// Test create, get, build, delete simple (non parametrized) job
func TestSimpleJobActions(t *testing.T) {
	var err error
	var jobName string = "test1"
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
	// Create job
	jobCreateResult := <-api.JobCreate(jobName, jobConfig)
	assert.NotNil(t, jobCreateResult)
	assert.NotNil(t, jobCreateResult.Response)
	assert.Nil(t, jobCreateResult.Error)
	assert.Equal(t, jobName, jobCreateResult.Response.DisplayName)
	assert.Equal(t, fmt.Sprintf("%s/job/%s/", baseURL, jobName), jobCreateResult.Response.URL)

	// Invoke build
	buildInvokeResult := <-api.BuildInvoke(jobName)
	assert.NotNil(t, buildInvokeResult)
	assert.Nil(t, buildInvokeResult.Error)
	assert.NotNil(t, buildInvokeResult.Response)
	invoked := buildInvokeResult.Response
	assert.NotNil(t, invoked.URL)
	assert.NotZero(t, invoked.ID)
	time.Sleep(20 * time.Second)

	// Get job information
	jobGetResult := <-api.JobGet(jobName)
	assert.NotNil(t, jobGetResult)
	assert.NotNil(t, jobGetResult.Response)
	assert.Nil(t, jobGetResult.Error)

	// Check some of common job information
	job := jobGetResult.Response
	assert.Equal(t, jobName, job.DisplayName)
	assert.Equal(t, fmt.Sprintf("%s/job/%s/", baseURL, jobName), job.URL)

	// Check some of build-related job information
	assert.False(t, job.InQueue)
	assert.True(t, job.LastBuild.Number == job.LastSuccessfulBuild.Number)
	assert.Zero(t, job.LastFailedBuild.Number)
	var expectedNextBuildNumber uint = 2
	assert.Equal(t, expectedNextBuildNumber, job.NextBuildNumber)

	// Get build by a known number for a particular job
	var jobBuildNumber uint = 1
	buildGetResult := <-api.BuildGetByNumber(jobName, jobBuildNumber)
	assert.NotNil(t, buildGetResult)
	assert.NotNil(t, buildGetResult.Response)
	assert.Nil(t, buildGetResult.Error)

	// Check some of build-related job information
	build := buildGetResult.Response
	assert.Equal(t, "SUCCESS", build.Result)
	assert.Equal(t, jobBuildNumber, build.Number)
	assert.False(t, build.Building)

	// Delete job
	err = <-api.JobDelete(jobName)
	assert.Nil(t, err)
}
