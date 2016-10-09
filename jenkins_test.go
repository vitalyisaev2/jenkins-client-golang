package jenkins_test

import (
	"log"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalyisaev2/jenkins-client-golang"
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
	api, err = jenkins.NewJenkins("http://localhost:8080", "admin", jenkinsAdminPassword, true)
	assert.NotNil(t, api)
	assert.Nil(t, err)

	result := <-api.RootInfo()
	assert.NotNil(t, result)
	assert.NotNil(t, result.Response)
	assert.NotEqual(t, 0, result.Response.NumExecutors)
	assert.Nil(t, result.Error)
}

func TestCreateJobAPI(t *testing.T) {
	jobName := "test1"
	jobConfig := []byte(`
<project>
  <keepDependencies>false</keepDependencies>
  <properties/>
  <scm class="hudson.scm.NullSCM"/>
  <canRoam>false</canRoam>
  <disabled>false</disabled>
  <blockBuildWhenDownstreamBuilding>false</blockBuildWhenDownstreamBuilding>
  <blockBuildWhenUpstreamBuilding>false</blockBuildWhenUpstreamBuilding>
  <triggers/>
  <concurrentBuild>false</concurrentBuild>
  <builders/>
  <publishers/>
  <buildWrappers/>
</project>
`)
	err := <-api.JobCreate(jobName, jobConfig)
	assert.Nil(t, err)

}
