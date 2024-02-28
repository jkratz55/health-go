package health

import (
	"fmt"
	"net/http"
)

type Status string

const (
	// StatusUp indicates the application is up and functioning as expected.
	StatusUp Status = "UP"
	// StatusDegraded indicates the application is up and functional but is
	// experiencing non-critical issues which is degrading the experience.
	StatusDegraded Status = "DEGRADED"
	// StatusDown indicates the application is not functional and for all intents
	// and purposes the application is down (not usable).
	StatusDown Status = "DOWN"
)

// HttpStatusCode returns the HTTP status code for the given status.
func (s Status) HttpStatusCode() int {
	switch s {
	case StatusUp:
		return http.StatusOK
	case StatusDegraded:
		// Although the service is degraded we still need to return an HTTP 2XX
		// status or Kubernetes will not consider the service available and not
		// route traffic to it.
		return http.StatusOK
	case StatusDown:
		return http.StatusServiceUnavailable
	default:
		// This can only happen by a programming error or someone trying to
		// skirt around the Status type and constants defined.
		panic(fmt.Errorf("%s is not a valid status", s))
	}
}
