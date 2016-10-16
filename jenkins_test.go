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
	baseURL            string = "http://localhost:8080"
	jenkinsAdminLogin  string = "admin"
	debug              bool   = true
	jobConfigWithSleep string = `
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
      <command>sleep 3;</command>
    </hudson.tasks.Shell>
  </builders>
  <publishers/>
  <buildWrappers/>
</project>
	`
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
	jobConfig := []byte(jobConfigWithSleep)
	// Create job
	jobCreateResult := <-api.JobCreate(jobName, jobConfig)
	assert.NotNil(t, jobCreateResult)
	assert.NotNil(t, jobCreateResult.Response)
	assert.Nil(t, jobCreateResult.Error)
	assert.Equal(t, jobName, jobCreateResult.Response.DisplayName)
	assert.Equal(t, fmt.Sprintf("%s/job/%s/", baseURL, jobName), jobCreateResult.Response.URL)

	// Check that Job exists, but is not enqueued or building
	jobExists := <-api.JobExists(jobName)
	assert.Nil(t, jobExists.Error)
	assert.True(t, jobExists.Response)
	jobInQueue := <-api.JobInQueue(jobName)
	assert.Nil(t, jobInQueue.Error)
	assert.False(t, jobInQueue.Response)
	jobIsBuilding := <-api.JobIsBuilding(jobName)
	assert.Nil(t, jobIsBuilding.Error)
	assert.False(t, jobIsBuilding.Response)

	// Invoke build
	buildInvokeResult := <-api.BuildInvoke(jobName)
	assert.NotNil(t, buildInvokeResult)
	assert.Nil(t, buildInvokeResult.Error)
	assert.NotNil(t, buildInvokeResult.Response)
	invoked := buildInvokeResult.Response
	assert.NotNil(t, invoked.URL)
	assert.NotZero(t, invoked.ID)

	// Wait until build will pass the queue and building process
	for {
		jobInQueue = <-api.JobInQueue(jobName)
		assert.Nil(t, jobInQueue.Error)
		if !jobInQueue.Response {
			fmt.Println("Job has passed the queue")
			break
		}
		fmt.Println("Job is in queue. Waiting for 1 sec...")
		time.Sleep(1 * time.Second)
	}

	for {
		jobIsBuilding = <-api.JobIsBuilding(jobName)
		assert.Nil(t, jobIsBuilding.Error)
		if !jobIsBuilding.Response {
			fmt.Println("Job has been built")
			break
		}
		fmt.Println("Job is building. Waiting for 1 sec...")
		time.Sleep(1 * time.Second)
	}

	// After
	//time.Sleep(15 * time.Second)

	// Get job information
	jobGetResult := <-api.JobGet(jobName, 0)
	assert.NotNil(t, jobGetResult)
	assert.NotNil(t, jobGetResult.Response)
	assert.Nil(t, jobGetResult.Error)

	// Check some of common job information
	job := jobGetResult.Response
	assert.Equal(t, jobName, job.DisplayName)
	assert.Equal(t, fmt.Sprintf("%s/job/%s/", baseURL, jobName), job.URL)

	// Check some of build-related job information
	assert.False(t, job.InQueue)
	assert.Equal(t, job.LastBuild.Number, job.LastSuccessfulBuild.Number)
	assert.Zero(t, job.LastFailedBuild.Number)
	var expectedNextBuildNumber uint = 2
	assert.Equal(t, expectedNextBuildNumber, job.NextBuildNumber)

	// Get build by a known number for a particular job
	var jobBuildNumber uint = 1
	buildGetResult := <-api.BuildGetByNumber(jobName, jobBuildNumber)
	assert.NotNil(t, buildGetResult)
	assert.NotNil(t, buildGetResult.Response)
	assert.Nil(t, buildGetResult.Error)

	// Get build by a known QueueID for a particular job
	buildGetByQueueIDResult := <-api.BuildGetByQueueID(jobName, invoked.ID)
	assert.NotNil(t, buildGetByQueueIDResult)
	assert.NotNil(t, buildGetByQueueIDResult.Response)
	assert.Nil(t, buildGetByQueueIDResult.Error)

	// Check that build information retrieved by this too method is equivalent
	assert.Equal(t, buildGetResult.Response, buildGetByQueueIDResult.Response)

	// Check some of build-related job information
	build := buildGetResult.Response
	assert.Equal(t, "SUCCESS", build.Result)
	assert.Equal(t, jobBuildNumber, build.Number)
	assert.False(t, build.Building)

	// Delete job
	err = <-api.JobDelete(jobName)
	assert.Nil(t, err)
}
