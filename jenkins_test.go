package jenkins_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vitalyisaev2/jenkins-client-golang"
)

func TestInit(t *testing.T) {
	api, err := jenkins.NewJenkins("http://localhost:8080", "admin", "5a37817dfc9e417887502aca337844e5")
	assert.NotNil(t, api)
	assert.Nil(t, err)

	result := <-api.RootInfo()
	assert.NotNil(t, result.Response)
	assert.Nil(t, result.Error)
}
