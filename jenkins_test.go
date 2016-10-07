package jenkins_test

import (
	"log"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalyisaev2/jenkins-client-golang"
)

var jenkinsAdminPassword string

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
	api, err := jenkins.NewJenkins("http://localhost:8080", "admin", jenkinsAdminPassword)
	assert.NotNil(t, api)
	assert.Nil(t, err)

	result := <-api.RootInfo()
	assert.NotNil(t, result)
	assert.NotNil(t, result.Response)
	assert.NotEqual(t, 0, result.Response.NumExecutors)
	assert.Nil(t, result.Error)
}
