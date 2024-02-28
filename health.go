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

type Health struct {
	components Components
}

func New(components ...Component) *Health {
	comps := make([]Component, 0)
	for _, c := range components {
		comps = append(comps, c)
	}
	return &Health{
		components: comps,
	}
}

func (h *Health) Register(component Component) {
	if component.Check == nil {
		panic("health: component must have a non-nil check")
	}
	h.components = append(h.components, component)
}

func (h *Health) Status(ctx context.Context) Status {
	return h.components.Status(ctx)
}

func (h *Health) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.components.ServeHTTP(w, r)
}

func (h *Health) HandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}
