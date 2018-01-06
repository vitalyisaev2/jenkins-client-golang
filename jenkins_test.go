// Package jenkins_test contains integrational tests for Jenkins client library;
// The only dependency is Jenkins itself running within Docker container with exposed port 8080.
package jenkins_test

import (
	"context"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/vitalyisaev2/jenkins-client-golang"
)

const (
	baseURL            string = "http://localhost:8080"
	login              string = "admin"
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

type jenkinsSuite struct {
	suite.Suite
	client jenkins.Client
	ctx    context.Context
}

func (s *jenkinsSuite) SetupSuite() {
	// Get admin temporary credentials for test purposes
	out, err := exec.Command("docker", "exec", "jenkins", "cat", "/var/jenkins_home/secrets/initialAdminPassword").Output()
	if !s.Assert().NoError(err) {
		s.FailNow("Couldn't get Docker admin password", err.Error())
	}
	password := string(out[:len(out)-1])

	s.client, err = jenkins.NewClient(baseURL, login, password, debug)
	s.Assert().NoError(err)
	s.Assert().NotNil(s.client)
}

// Test API initialisation
func (s *jenkinsSuite) TestRootInfo() {
	info, err := s.client.RootInfo(s.ctx)
	s.Assert().NoError(err)
	s.Assert().NotNil(info)
	s.Assert().NotZero(info.NumExecutors)
}

// Test create, get, build, delete simple (non parametrized) job
func (s *jenkinsSuite) TestSimpleJobActions() {
	var (
		err  error
		name string = "test1"
	)

	// Create job
	jobCreated, err := s.client.JobCreate(s.ctx, name, jobConfigWithSleep)
	s.Assert().NotNil(jobCreated)
	s.Assert().NoError(err)
	s.Assert().Equal(name, jobCreated.DisplayName)

	// Check that Job exists, but is not enqueued or building
	exists, err := s.client.JobExists(s.ctx, name)
	s.Assert().NoError(err)
	s.Assert().True(exists)

	enqueued, err := s.client.JobInQueue(s.ctx, name)
	s.Assert().NoError(err)
	s.Assert().False(enqueued)

	building, err := s.client.JobIsBuilding(s.ctx, name)
	s.Assert().NoError(err)
	s.Assert().False(building)

	// Invoke build
	invoked, err := s.client.BuildInvoke(s.ctx, name)
	s.Assert().NoError(err)
	s.Assert().NotNil(invoked.URL)
	s.Assert().NotZero(invoked.ID)

	// Poll API until build will pass the queue (1) and building process (2)
	for {
		enqueued, err := s.client.JobInQueue(s.ctx, name)
		s.Assert().NoError(err)
		if !enqueued {
			s.T().Log("Job has passed the queue")
			break
		}
		s.T().Log("Job is in queue. Waiting for 1 sec...")
		time.Sleep(1 * time.Second)
	}

	for {
		building, err = s.client.JobIsBuilding(s.ctx, name)
		s.Assert().NoError(err)
		if !building {
			fmt.Println("Job has been built")
			break
		}
		fmt.Println("Job is building. Waiting for 1 sec...")
		time.Sleep(1 * time.Second)
	}

	// Get and check job information
	jobObtained, err := s.client.JobGet(s.ctx, name, 0)
	s.Assert().NoError(err)
	s.Assert().NotNil(jobObtained)
	s.Assert().Equal(name, jobObtained.DisplayName)

	// Check some build-related job information
	s.Assert().False(jobObtained.InQueue)
	s.Assert().Equal(jobObtained.LastBuild.Number, jobObtained.LastSuccessfulBuild.Number)
	s.Assert().Zero(jobObtained.LastFailedBuild.Number)
	var expectedNextBuildNumber int = 2
	s.Assert().Equal(expectedNextBuildNumber, jobObtained.NextBuildNumber)

	// Get build by a known number for a particular job
	var jobBuildNumber uint = 1
	buildGetResult := s.client.BuildGetByNumber(jobName, jobBuildNumber)
	s.Assert().NotNil(buildGetResult)
	s.Assert().NotNil(buildGetResult.Response)
	s.Assert().Nil(buildGetResult.Error)

	// Get build by a known QueueID for a particular job
	buildGetByQueueIDResult := s.client.BuildGetByQueueID(jobName, invoked.ID)
	s.Assert().NotNil(buildGetByQueueIDResult)
	s.Assert().NotNil(buildGetByQueueIDResult.Response)
	s.Assert().Nil(buildGetByQueueIDResult.Error)

	// Check that build information retrieved by this too method is equivalent
	s.Assert().Equal(buildGetResult.Response, buildGetByQueueIDResult.Response)

	// Check some of build-related job information
	build := buildGetResult.Response
	s.Assert().Equal("SUCCESS", build.Result)
	s.Assert().Equal(jobBuildNumber, build.Number)
	s.Assert().False(build.Building)

	// Delete job
	err = s.client.JobDelete(jobName)
	s.Assert().Nil(err)
}

/*
func TestComplicatedUnmarshalling(t *testing.T) {
	s := `{"_class":"hudson.model.FreeStyleProject","actions":[{"_class":"hudson.model.ParametersDefinitionProperty","parameterDefinitions":[{"_class":"hudson.model.BooleanParameterDefinition","defaultParameterValue":{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true},"description":"","name":"x","type":"BooleanParameterDefinition"}]},{"_class":"com.cloudbees.plugins.credentials.ViewCredentialsAction","stores":{}}],"description":"","displayName":"BuildControl_A","displayNameOrNull":null,"name":"BuildControl_A","url":"http://kiki:8080/job/BuildControl_A/","buildable":true,"builds":[{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.ParametersAction","parameters":[{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true}]},{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#5","duration":5017,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #5","id":"5","keepLog":false,"number":5,"queueId":7,"result":"SUCCESS","timestamp":1476801392270,"url":"http://kiki:8080/job/BuildControl_A/5/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.ParametersAction","parameters":[{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true}]},{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#4","duration":5034,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #4","id":"4","keepLog":false,"number":4,"queueId":6,"result":"FAILURE","timestamp":1476801340504,"url":"http://kiki:8080/job/BuildControl_A/4/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.ParametersAction","parameters":[{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true}]},{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#3","duration":5012,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #3","id":"3","keepLog":false,"number":3,"queueId":5,"result":"FAILURE","timestamp":1476801224113,"url":"http://kiki:8080/job/BuildControl_A/3/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#2","duration":5017,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #2","id":"2","keepLog":false,"number":2,"queueId":4,"result":"FAILURE","timestamp":1476801188749,"url":"http://kiki:8080/job/BuildControl_A/2/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#1","duration":5012,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #1","id":"1","keepLog":false,"number":1,"queueId":3,"result":"FAILURE","timestamp":1476800867225,"url":"http://kiki:8080/job/BuildControl_A/1/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]}],"color":"blue","firstBuild":{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#1","duration":5012,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #1","id":"1","keepLog":false,"number":1,"queueId":3,"result":"FAILURE","timestamp":1476800867225,"url":"http://kiki:8080/job/BuildControl_A/1/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},"healthReport":[{"description":"Build stability: 4 out of the last 5 builds failed.","iconClassName":"icon-health-00to19","iconUrl":"health-00to19.png","score":20}],"inQueue":false,"keepDependencies":false,"lastBuild":{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.ParametersAction","parameters":[{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true}]},{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#5","duration":5017,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #5","id":"5","keepLog":false,"number":5,"queueId":7,"result":"SUCCESS","timestamp":1476801392270,"url":"http://kiki:8080/job/BuildControl_A/5/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},"lastCompletedBuild":{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.ParametersAction","parameters":[{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true}]},{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#5","duration":5017,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #5","id":"5","keepLog":false,"number":5,"queueId":7,"result":"SUCCESS","timestamp":1476801392270,"url":"http://kiki:8080/job/BuildControl_A/5/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},"lastFailedBuild":{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.ParametersAction","parameters":[{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true}]},{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#4","duration":5034,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #4","id":"4","keepLog":false,"number":4,"queueId":6,"result":"FAILURE","timestamp":1476801340504,"url":"http://kiki:8080/job/BuildControl_A/4/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},"lastStableBuild":{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.ParametersAction","parameters":[{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true}]},{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#5","duration":5017,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #5","id":"5","keepLog":false,"number":5,"queueId":7,"result":"SUCCESS","timestamp":1476801392270,"url":"http://kiki:8080/job/BuildControl_A/5/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},"lastSuccessfulBuild":{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.ParametersAction","parameters":[{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true}]},{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#5","duration":5017,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #5","id":"5","keepLog":false,"number":5,"queueId":7,"result":"SUCCESS","timestamp":1476801392270,"url":"http://kiki:8080/job/BuildControl_A/5/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},"lastUnstableBuild":null,"lastUnsuccessfulBuild":{"_class":"hudson.model.FreeStyleBuild","actions":[{"_class":"hudson.model.ParametersAction","parameters":[{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true}]},{"_class":"hudson.model.CauseAction","causes":[{"_class":"hudson.model.Cause$UserIdCause","shortDescription":"Started by user admin","userId":"admin","userName":"admin"}]}],"artifacts":[],"building":false,"description":null,"displayName":"#4","duration":5034,"estimatedDuration":5021,"executor":null,"fullDisplayName":"BuildControl_A #4","id":"4","keepLog":false,"number":4,"queueId":6,"result":"FAILURE","timestamp":1476801340504,"url":"http://kiki:8080/job/BuildControl_A/4/","builtOn":"","changeSet":{"_class":"hudson.scm.EmptyChangeLogSet","items":[],"kind":null},"culprits":[]},"nextBuildNumber":6,"property":[{"_class":"hudson.model.ParametersDefinitionProperty","parameterDefinitions":[{"_class":"hudson.model.BooleanParameterDefinition","defaultParameterValue":{"_class":"hudson.model.BooleanParameterValue","name":"x","value":true},"description":"","name":"x","type":"BooleanParameterDefinition"}]}],"queueItem":null,"concurrentBuild":false,"downstreamProjects":[],"scm":{"_class":"hudson.scm.NullSCM","browser":null,"type":"hudson.scm.NullSCM"},"upstreamProjects":[]}`
	var receiver response.Job
	buf := bytes.NewBufferString(s)
	err := json.NewDecoder(buf).Decode(&receiver)
	if ute, ok := err.(*json.UnmarshalTypeError); ok {
		fmt.Printf("UnmarshalTypeError %v - %v - %v", ute.Value, ute.Type, ute.Offset)
	}
	s.Assert().Nil(err)
}
*/

func (s *jenkinsSuite) TearDownSuite() {}

func TestJenkins(t *testing.T) {
	s := &jenkinsSuite{
		ctx: context.Background(),
	}
	suite.Run(t, s)
}
