package rundeck

import "net/http"

// RundeckTransport is a custom HTTP transport for Rundeck API requests.
type RundeckTransport struct {
	Transport http.RoundTripper
	AuthToken string
}

// RoundTrip adds the auth header before sending the request.
func (c *RundeckTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("X-Rundeck-Auth-Token", c.AuthToken)

	transport := c.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	// Proceed with the request
	return transport.RoundTrip(req)
}
