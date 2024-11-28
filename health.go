package health

import (
	"context"
	"net/http"
	"time"
)

var startTimestamp time.Time

func init() {
	startTimestamp = time.Now()
}

// Health monitors the health of services/components that an application depends
// on and tracks their status to determine the overall health of the application.
//
// Health implements the http.Handler interface and can be used to expose the
// health status of the application via an HTTP endpoint.
type Health struct {
	components Components
	ctx        context.Context
	cancel     context.CancelFunc
}

// New initializes a new Health instance with the provided components.
//
// Additional components can be registered by calling the Register method on the
// Health instance.
func New(components ...Component) *Health {
	comps := make([]*Component, 0)
	for _, c := range components {
		comp := c
		comps = append(comps, &comp)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Health{
		components: comps,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Register adds a component to be monitored and considered in the overall health
// of the application.
//
// Panics if the component does not have a non-nil check function.
func (h *Health) Register(component Component) {
	if component.Check == nil {
		panic("health: component must have a non-nil check")
	}
	component.init()
	go component.monitor(h.ctx)
	h.components = append(h.components, &component)
}

// Status returns the overall status of the application.
func (h *Health) Status(ctx context.Context) Status {
	return h.components.Status(ctx)
}

// ServeHTTP is the HTTP handler for the health endpoint which returns the overall
// health status of the application along with the status of each component. If the
// application overall status is Up or Degraded a 200 OK status code is returned.
// If the application overall status is Down a 503 Service Unavailable status code
// is returned.
func (h *Health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.components.ServeHTTP(w, r)
}

// HandlerFunc returns an http.HandlerFunc for the health endpoint which returns
// the overall health status of the application along with the status of each
// component. If the application overall status is Up or Degraded a 200 OK status
// code is returned. If the application overall status is Down a 503 Service
// Unavailable status code is returned.
func (h *Health) HandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

// Shutdown stops monitoring the health of the components but will continue to
// return the last known status of the components.
func (h *Health) Shutdown() {
	h.cancel()
}
