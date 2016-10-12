package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

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

	// Select dump method and run it
	switch method {
	case ResponseDumpNone:
		return nil
	case ResponseDumpDefaultJSON:
		return dm.defaultJSON(httpResponse, receiver)
	case ResponseDumpHeaderLocation:
		// Cast receiver to URL
		if receiverURL, casted := receiver.(*url.URL); !casted {
			return fmt.Errorf("Cannot cast receiver to *url.URL")
		} else {
			return dm.headerLocation(httpResponse, receiverURL)
		}
	default:
		return fmt.Errorf("Unknown ResponseDumpMethod")
	}
}

// Unmarshal location header to a given URL
func (dm *dumper) headerLocation(httpResponse *http.Response, receiver *url.URL) error {

	// Check response status
	switch httpResponse.StatusCode {
	case http.StatusCreated:
		break
	default:
		return fmt.Errorf("Bad response status: %s", httpResponse.Status)
	}

	location, err := httpResponse.Location()
	//fmt.Printf("receiver %v (%T) (%p)\n", receiver, receiver, receiver)
	if err != nil {
		return err
	} else {
		*receiver = *location
		//fmt.Printf("receiver %v (%T) (%p)\n", receiver, receiver, receiver)
		return nil
	}
}

// Unmarshal JSON to a given receiver
func (dm *dumper) defaultJSON(httpResponse *http.Response, receiver result.Result) error {

	// Check response status
	switch httpResponse.StatusCode {
	case http.StatusOK:
		break
	default:
		return fmt.Errorf("Bad response status: %s", httpResponse.Status)
	}

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
