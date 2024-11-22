package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Component represents a single component that can be checked for health.
type Component struct {

	// Name or identifier of the component. Each component should have a unique
	// name. Otherwise, the response from ServeHTTP will be not be particularly
	// helpful in identifying which component is in a bad state.
	Name string

	// Determines if the component is critical to the overall health and
	// functionality of the system. If a component is marked as critical, and
	// it fails its check, the overall status of the system will be down.
	Critical bool

	// Timeout for the health check. If the health check takes longer than the
	// timeout, the check is considered to have failed. The default value is
	// 5 seconds.
	Timeout time.Duration

	// Interval between health checks. If the interval is set to zero, a default
	// interval of 15 seconds will be used. The Interval must be greater than
	// the Timeout.
	Interval time.Duration

	// The health check of the component.
	//
	// A nil Check will cause a panic.
	Check CheckFunc

	status Status
}

func (c *Component) init() {
	if strings.TrimSpace(c.Name) == "" {
		c.Name = "MISSING NAME"
	}
	if c.Timeout == 0 {
		c.Timeout = 5 * time.Second
	}
	if c.Interval == 0 {
		c.Interval = 15 * time.Second
	}
	if c.Timeout >= c.Interval {
		c.Interval = c.Timeout + 1*time.Second
		fmt.Println("Timeout was greater than or equal to interval. Setting interval to timeout + 1 second.")
	}
	c.status = StatusUp
}

// monitor performs a healthcheck on the component at regular intervals and
// updates the status of the component.
func (c *Component) monitor(ctx context.Context) {
	ticker := time.NewTicker(c.Interval)
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
			err := c.Check(ctx)
			cancel()

			// If the health check fails, the status of the component is set to
			// down, otherwise it is set to up.
			if err != nil {
				c.status = StatusDown
			} else {
				c.status = StatusUp
			}
		case <-ctx.Done():
			return
		}
	}
}

// ComponentStatus represents the status of a component.
type ComponentStatus struct {
	Name     string `json:"name"`
	Critical bool   `json:"critical"`
	Status   Status `json:"status"`
}

// Components is a collection of components that can be checked for health.
//
// Components implements http.Handler and can be used to serve as a readiness
// or liveness health check endpoint.
type Components []*Component

// Status returns the overall status of the components.
func (c Components) Status(ctx context.Context) Status {
	status := StatusUp
	for _, component := range c {
		// If the component is critical, and it's down, the overall status is down.
		if component.status == StatusDown && component.Critical {
			status = StatusDown
		}
		// If the component is not critical, and it's down, the overall status
		// is degraded is set to degraded unless the overall status has already
		// been determined to be down.
		if component.status == StatusDown && !component.Critical && status != StatusDown {
			status = StatusDegraded
		}
		// If the component is degraded regardless of its criticality, the overall
		// status shall be considered degraded unless the overall status has already
		// been determined to be down.
		if component.status == StatusDegraded && status != StatusDown {
			status = StatusDegraded
		}
	}
	return status
}

// ComponentStatus returns the status of each component.
func (c Components) ComponentStatus(ctx context.Context) []ComponentStatus {
	statuses := make([]ComponentStatus, 0, len(c))
	for _, component := range c {
		statuses = append(statuses, ComponentStatus{
			Name:     component.Name,
			Critical: component.Critical,
			Status:   component.status,
		})
	}
	return statuses
}

func (c Components) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	type statusResponse struct {
		Status     Status            `json:"status"`
		Uptime     string            `json:"uptime"`
		Components []ComponentStatus `json:"components"`
	}

	resp := make([]ComponentStatus, 0, len(c))

	status := StatusUp
	for _, component := range c {
		// If the component is critical, and it's down, the overall status is down.
		if component.status == StatusDown && component.Critical {
			status = StatusDown
		}
		// If the component is not critical, and it's down, the overall status
		// is degraded is set to degraded unless the overall status has already
		// been determined to be down.
		if component.status == StatusDown && !component.Critical && status != StatusDown {
			status = StatusDegraded
		}
		// If the component is degraded regardless of its criticality, the overall
		// status shall be considered degraded unless the overall status has already
		// been determined to be down.
		if component.status == StatusDegraded && status != StatusDown {
			status = StatusDegraded
		}

		resp = append(resp, ComponentStatus{
			Name:     component.Name,
			Critical: component.Critical,
			Status:   component.status,
		})
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(status.HttpStatusCode())

	_ = json.NewEncoder(w).Encode(statusResponse{
		Status:     status,
		Uptime:     time.Since(startTimestamp).String(),
		Components: resp,
	})
}
