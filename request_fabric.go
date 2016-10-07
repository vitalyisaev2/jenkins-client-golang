package jenkins

import (
	"fmt"
	"io"
	"net/http"
)

type requestFabric struct {
	baseURL  string
	username string
	password string
}

func (rb *requestFabric) newJSONRequest(method string, route string, body io.Reader) (*http.Request, error) {
	URL := fmt.Sprintf("%s%s/api/json", rb.baseURL, route)
	//fmt.Println(URL)

	req, err := http.NewRequest(method, URL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(rb.username, rb.password)

	return req, nil
}
