package checks

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HttpClient is an interface for making HTTP requests.
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// HttpCheckConfig is the configuration for an HTTP check.
type HttpCheckConfig struct {

	// Http method to use for the request.
	Method string

	// Url to check for the health check.
	Url string

	// Timeout for the request. If the request takes longer than the timeout,
	// the check is considered to have failed. The default value is 3 seconds.
	Timeout time.Duration

	// HttpClient to use for the request. If nil, http.DefaultClient is used.
	HttpClient HttpClient
}

// HttpCheck returns a CheckFunc that checks the health of an HTTP endpoint.
// The endpoint is considered healthy if it returns a 2xx status code.
func HttpCheck(conf HttpCheckConfig) func(ctx context.Context) error {
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

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("unnsuccessful http status code %d", resp.StatusCode)
		}

		return nil
	}
}
