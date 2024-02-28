package health

import (
	"context"
	"encoding/json"
	"net/http"
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

	// The health check of the component.
	//
	// A nil Check will cause a panic.
	Check CheckFunc
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
type Components []Component

// Status returns the overall status of the components.
func (c Components) Status(ctx context.Context) Status {
	status := StatusUp
	for _, component := range c {
		cs := component.Check(ctx)

		// If the component is critical, and it's down, the overall status is down.
		if cs == StatusDown && component.Critical {
			status = StatusDown
		}
		// If the component is not critical, and it's down, the overall status
		// is degraded is set to degraded unless the overall status has already
		// been determined to be down.
		if cs == StatusDown && !component.Critical && status != StatusDown {
			status = StatusDegraded
		}
		// If the component is degraded regardless of its criticality, the overall
		// status shall be considered degraded unless the overall status has already
		// been determined to be down.
		if cs == StatusDegraded && status != StatusDown {
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
			Status:   component.Check(ctx),
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
		cs := component.Check(r.Context())

		// If the component is critical, and it's down, the overall status is down.
		if cs == StatusDown && component.Critical {
			status = StatusDown
		}
		// If the component is not critical, and it's down, the overall status
		// is degraded is set to degraded unless the overall status has already
		// been determined to be down.
		if cs == StatusDown && !component.Critical && status != StatusDown {
			status = StatusDegraded
		}
		// If the component is degraded regardless of its criticality, the overall
		// status shall be considered degraded unless the overall status has already
		// been determined to be down.
		if cs == StatusDegraded && status != StatusDown {
			status = StatusDegraded
		}

		resp = append(resp, ComponentStatus{
			Name:     component.Name,
			Critical: component.Critical,
			Status:   cs,
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
