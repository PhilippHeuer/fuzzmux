package rundeck

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Job struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Group           string `json:"group"`
	Project         string `json:"project"`
	Description     string `json:"description"`
	Href            string `json:"href"`
	Permalink       string `json:"permalink"`
	Scheduled       bool   `json:"scheduled"`
	ScheduleEnabled bool   `json:"scheduleEnabled"`
	Enabled         bool   `json:"enabled"`
}

// Client is a simple HTTP client to interact with the rundeck API
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient initializes a new API client
func NewClient(baseURL string, accessToken string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Transport: &RundeckTransport{
				Transport: http.DefaultTransport,
				AuthToken: accessToken,
			},
		},
	}
}

// GetJobs fetches jobs for a given project
func (c *Client) GetJobs(project string, queryParams map[string]string) ([]Job, error) {
	endpoint := fmt.Sprintf("%s/api/14/project/%s/jobs", c.BaseURL, project)
	reqURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	query := reqURL.Query()
	for key, value := range queryParams {
		if value != "" {
			query.Add(key, value)
		}
	}
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %v", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var jobs []Job
	if err = json.NewDecoder(resp.Body).Decode(&jobs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return jobs, nil
}
