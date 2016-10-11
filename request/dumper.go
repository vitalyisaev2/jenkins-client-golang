package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/vitalyisaev2/jenkins-client-golang/result"
)

// ResponseDumpMethod allows to specify how exactly response from Jenkins API will be dumped to receiver
type ResponseDumpMethod uint

const (
	// Omit dumping
	ResponseDumpNone ResponseDumpMethod = iota
	// ResponseDumpDefaultJSON unmarshalles JSON into a given struct
	ResponseDumpDefaultJSON
	// ResponseDumpHeaderLocation dumps Location response header
	ResponseDumpHeaderLocation
)

// Jenkins API may answer you in many different ways;
// this object holds collection of dumping functions
type dumper struct {
	debug bool
}

func (dm *dumper) dump(httpResponse *http.Response, receiver result.Result, method ResponseDumpMethod) error {

	// Check response status
	switch httpResponse.StatusCode {
	case http.StatusOK, http.StatusCreated:
		break
	default:
		return fmt.Errorf("%s", httpResponse.Status)
	}

	// Select dump method and run it
	switch method {
	case ResponseDumpNone:
		return nil
	case ResponseDumpDefaultJSON:
		return dm.defaultJSON(httpResponse, receiver)
	case ResponseDumpHeaderLocation:
		return fmt.Errorf("Not implemented yet")
	default:
		return fmt.Errorf("Unknown ResponseDumpMethod")
	}
}

func (dm *dumper) defaultJSON(httpResponse *http.Response, receiver result.Result) error {
	var err error

	defer httpResponse.Body.Close()
	switch dm.debug {
	case true:
		// FIXME: use logger
		{
			dumpedBody, _ := ioutil.ReadAll(httpResponse.Body)
			dumpedBodyReader := bytes.NewBuffer(dumpedBody)
			fmt.Printf("ResponseBody: %s\n", string(dumpedBody))
			switch receiver {
			case nil:
				return nil
			default:
				err = json.NewDecoder(dumpedBodyReader).Decode(receiver)
			}
		}
	case false:
		{
			switch receiver {
			case nil:
				return nil
			default:
				err = json.NewDecoder(httpResponse.Body).Decode(receiver)
			}
		}
	}
	return err
}
