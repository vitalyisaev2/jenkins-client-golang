package jenkins_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vitalyisaev2/jenkins-client-golang"
	"github.com/vitalyisaev2/jenkins-client-golang/response"
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
	jobConfig := bytes.NewBufferString(jobConfigWithSleep)

	// Create job
	jobCreateResult := <-api.JobCreate(jobName, jobConfig)
	assert.NotNil(t, jobCreateResult)
	assert.NotNil(t, jobCreateResult.Response)
	assert.Nil(t, jobCreateResult.Error)
	assert.Equal(t, jobName, jobCreateResult.Response.DisplayName)

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

	// Get job information
	jobGetResult := <-api.JobGet(jobName, 0)
	assert.NotNil(t, jobGetResult)
	assert.NotNil(t, jobGetResult.Response)
	assert.Nil(t, jobGetResult.Error)

	// Check some of common job information
	job := jobGetResult.Response
	assert.Equal(t, jobName, job.DisplayName)

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

func TestComplicatedUnmarshalling(t *testing.T) {
	s := `{"_class":"hudson.model.FreeStyleProject","actions":[{"_class":"hudson.model.ParametersDefinitionProperty","parameterDefinitions":[{"_class":"hudson.model.BooleanParameterDefinition","defaultParameterValue":{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true},"description":"","name":"x","type":"BooleanParameterDefinition"}]},{"_class":"com.cloudbees.plugins.credentials.ViewCredentialsAction","stores":{}}],"description":"","displayName":"BuildControl_A","displayNameOrNull":null,"name":"BuildControl_A","url":"http://kiki:8080/job/BuildControl_A/","buildable":true,"builds":[{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.ParametersAction","parameters":[{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true}]},{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#5","duration":5017,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #5","id":"5","keepLog":false,"number":5,"queueId":7,"result":"SUCCESS","timestamp":1476801392270,"url":"http://kiki:8080/job/BuildControl_A/5/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.ParametersAction","parameters":[{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true}]},{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#4","duration":5034,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #4","id":"4","keepLog":false,"number":4,"queueId":6,"result":"FAILURE","timestamp":1476801340504,"url":"http://kiki:8080/job/BuildControl_A/4/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.ParametersAction","parameters":[{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true}]},{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#3","duration":5012,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #3","id":"3","keepLog":false,"number":3,"queueId":5,"result":"FAILURE","timestamp":1476801224113,"url":"http://kiki:8080/job/BuildControl_A/3/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#2","duration":5017,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #2","id":"2","keepLog":false,"number":2,"queueId":4,"result":"FAILURE","timestamp":1476801188749,"url":"http://kiki:8080/job/BuildControl_A/2/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#1","duration":5012,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #1","id":"1","keepLog":false,"number":1,"queueId":3,"result":"FAILURE","timestamp":1476800867225,"url":"http://kiki:8080/job/BuildControl_A/1/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]}],"color":"blue","firstBuild":{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#1","duration":5012,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #1","id":"1","keepLog":false,"number":1,"queueId":3,"result":"FAILURE","timestamp":1476800867225,"url":"http://kiki:8080/job/BuildControl_A/1/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},"healthReport":[{"description":"Build stability: 4 out of the last 5 builds failed.","iconClassName":"icon-health-00to19","iconUrl":"health-00to19.png","score":20}],"inQueue":false,"keepDependencies":false,"lastBuild":{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.ParametersAction","parameters":[{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true}]},{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#5","duration":5017,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #5","id":"5","keepLog":false,"number":5,"queueId":7,"result":"SUCCESS","timestamp":1476801392270,"url":"http://kiki:8080/job/BuildControl_A/5/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},"lastCompletedBuild":{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.ParametersAction","parameters":[{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true}]},{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#5","duration":5017,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #5","id":"5","keepLog":false,"number":5,"queueId":7,"result":"SUCCESS","timestamp":1476801392270,"url":"http://kiki:8080/job/BuildControl_A/5/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},"lastFailedBuild":{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.ParametersAction","parameters":[{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true}]},{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#4","duration":5034,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #4","id":"4","keepLog":false,"number":4,"queueId":6,"result":"FAILURE","timestamp":1476801340504,"url":"http://kiki:8080/job/BuildControl_A/4/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},"lastStableBuild":{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.ParametersAction","parameters":[{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true}]},{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#5","duration":5017,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #5","id":"5","keepLog":false,"number":5,"queueId":7,"result":"SUCCESS","timestamp":1476801392270,"url":"http://kiki:8080/job/BuildControl_A/5/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},"lastSuccessfulBuild":{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.ParametersAction","parameters":[{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true}]},{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#5","duration":5017,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #5","id":"5","keepLog":false,"number":5,"queueId":7,"result":"SUCCESS","timestamp":1476801392270,"url":"http://kiki:8080/job/BuildControl_A/5/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},"lastUnstableBuild":null,"lastUnsuccessfulBuild":{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.ParametersAction","parameters":[{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true}]},{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#4","duration":5034,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #4","id":"4","keepLog":false,"number":4,"queueId":6,"result":"FAILURE","timestamp":1476801340504,"url":"http://kiki:8080/job/BuildControl_A/4/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},"nextBuildNumber":6,"property":[{"_class":"hudson.model.ParametersDefinitionProperty","parameterDefinitions":[{"_class":"hudson.model.BooleanParameterDefinition","defaultParameterValue":{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true},"description":"","name":"x","type":"BooleanParameterDefinition"}]}],"queueItem":null,"concurrentBuild":false,"downstreamProjects":[],"scm":{"_class":"hudson.scm.NullSCM","browser":null,"type":"hudson.scm.NullSCM"},"upstreamProjects":[]}`
	var receiver response.Job
	buf := bytes.NewBufferString(s)
	err := json.NewDecoder(buf).Decode(&receiver)
	if ute, ok := err.(*json.UnmarshalTypeError); ok {
		fmt.Printf("UnmarshalTypeError %v - %v - %v", ute.Value, ute.Type, ute.Offset)
	}
	assert.Nil(t, err)
}
