package xmlrpc

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
)

// Client implements a basic XMLRPC client
type Client struct {
	addr       string
	httpClient *http.Client
	username   string
	password   string
}

// NewClient returns a new instance of Client
// Pass in a true value for `insecure` to turn off certificate verification
func NewClient(addr string, insecure bool) *Client {
	transport := &http.Transport{}
	if insecure {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	httpClient := &http.Client{Transport: transport}

	return &Client{
		addr:       addr,
		httpClient: httpClient,
	}
}

// SetBasicAuth sets the request's HTTP basic authentication
func (c *Client) SetBasicAuth(username string, password string) {
	c.username = username
	c.password = password
}

// Call calls the method with "name" with the given args
// Returns the result, and an error for communication errors
func (c *Client) Call(name string, args ...interface{}) (interface{}, error) {
	postBody := bytes.NewBuffer(nil)
	e := Marshal(postBody, name, args...)
	if e != nil {
		return nil, e
	}

	req, e := http.NewRequest(http.MethodPost, c.addr, postBody)
	if e != nil {
		return nil, e
	}

	req.Header.Set("Content-Type", "text/xml")

	if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	r, e := c.httpClient.Do(req)
	if e != nil {
		return nil, e
	}
	defer r.Body.Close()

	_, v, f, e := Unmarshal(r.Body)
	if f != nil {
		e = fmt.Errorf("Error: %v: %v", e, f)
	}
	return v, e
}
