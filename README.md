[![Build Status](https://travis-ci.org/vitalyisaev2/jenkins-client-golang.svg?branch=master)](https://travis-ci.org/vitalyisaev2/jenkins-client-golang)
[![Coverage Status](https://coveralls.io/repos/github/vitalyisaev2/jenkins-client-golang/badge.svg)](https://coveralls.io/github/vitalyisaev2/jenkins-client-golang)

This library is inspired by [bndr/gojenkins](https://github.com/bndr/gojenkins) but tends to be goroutine-safe. Not operational yet.

### Examples
#### API initialization
```go
import "github.com/vitalyisaev2/jenkins-client-golang"

// API required parameters
url := "http://localhost:8080/"
login := "login"
password := "password"
// Use true/false in case if you want additional debug information to be enabled/disabled
debug := false

api, err := jenkins.NewJenkins(url, login, password, debug)
if err != nil {
    return err
}
```
For more examples please look through source code of [jenkins_test.go](https://github.com/vitalyisaev2/jenkins-client-golang/blob/master/jenkins_test.go).
