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
	var jobBuildNumber int = 1
	buildByNumber, err := s.client.BuildGetByNumber(s.ctx, name, jobBuildNumber)
	s.Assert().NoError(err)
	s.Assert().NotNil(buildByNumber)

	// Get build by a known QueueID for a particular job
	buildByQueueID, err := s.client.BuildGetByQueueID(s.ctx, name, invoked.ID)
	s.Assert().NoError(err)
	s.Assert().NotNil(buildByQueueID)

	// Check that build information retrieved by this too method is equivalent
	s.Assert().Equal(buildByNumber, buildByQueueID)

	// Check some of build-related job information
	s.Assert().Equal("SUCCESS", buildByNumber.Result)
	s.Assert().Equal(jobBuildNumber, buildByNumber.Number)
	s.Assert().False(buildByNumber.Building)

	// Delete job
	err = s.client.JobDelete(s.ctx, name)
	s.Assert().Nil(err)
}

func (s *jenkinsSuite) TearDownSuite() {}

func TestJenkins(t *testing.T) {
	s := &jenkinsSuite{
		ctx: context.Background(),
	}
	suite.Run(t, s)
}
