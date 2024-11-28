package http

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Client is an interface for making HTTP requests.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// Config is the configuration for an HTTP check.
type Config struct {

	// Http method to use for the request.
	Method string

	// Url to check for the health check.
	Url string

	// Timeout for the request. If the request takes longer than the timeout,
	// the check is considered to have failed. The default value is 3 seconds.
	Timeout time.Duration

	// ExpectedStatusCode is the status code that the endpoint is expected to
	// return to be considered healthy. The default zero-value will configure
	// the check to expect a 2xx status code, meaning it doesn't expect a
	// specific status code, just a successful one.
	ExpectedStatusCode int

	// HttpClient to use for the request. If nil, http.DefaultClient is used.
	HttpClient Client
}

// New returns a CheckFunc that checks the health of an HTTP endpoint.
// The endpoint is considered healthy if it returns a 2xx status code.
func New(conf Config) func(ctx context.Context) error {
	if conf.Timeout == 0 {
		conf.Timeout = 3 * time.Second
	}
	if conf.HttpClient == nil {
		conf.HttpClient = http.DefaultClient
	}
	return func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(ctx, conf.Method, conf.Url, nil)
		if err != nil {
			return fmt.Errorf("failed to create http request: %w", err)
		}

		ctx, cancel := context.WithTimeout(ctx, conf.Timeout)
		defer cancel()

		// Inform the server it doesn't need to keep the connection open.
		req.Header.Set("Connection", "close")

		resp, err := conf.HttpClient.Do(req)
		if err != nil {
			return fmt.Errorf("http client failed: %w", err)
		}
		defer resp.Body.Close()

		// If an expected status code is not configured then we expect a 2xx
		// status code. Anything outside of that range is considered a failure.
		if conf.ExpectedStatusCode == 0 {
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return fmt.Errorf("unnsuccessful http status code %d", resp.StatusCode)
			}
		}
		// If an expected status code is configured then we expect that status
		// code to be returned. Anything else is considered a failure.
		if conf.ExpectedStatusCode != 0 && resp.StatusCode != conf.ExpectedStatusCode {
			return fmt.Errorf("unexpected http status code %d, expected %d", resp.StatusCode, conf.ExpectedStatusCode)
		}

		return nil
	}
}
