package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name               string
		init               func() *Health
		expectedComponents int
	}{
		{
			name: "New with No Components",
			init: func() *Health {
				return New()
			},
			expectedComponents: 0,
		},
		{
			name: "New With Components",
			init: func() *Health {
				components := make([]Component, 0)
				components = append(components, Component{
					Name:     "redis",
					Critical: false,
					Check: func(ctx context.Context) Status {
						return StatusUp
					},
				})
				components = append(components, Component{
					Name:     "mongo",
					Critical: true,
					Check: func(ctx context.Context) Status {
						return StatusUp
					},
				})
				return New(components...)
			},
			expectedComponents: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.init()
			assert.Equal(t, tt.expectedComponents, len(actual.components))
		})
	}
}

func TestHealth_Register(t *testing.T) {
	h := New()
	h.Register(Component{
		Name:     "redis",
		Critical: false,
		Check: func(ctx context.Context) Status {
			return StatusUp
		},
	})
	h.Register(Component{
		Name:     "mongo",
		Critical: true,
		Check: func(ctx context.Context) Status {
			return StatusUp
		},
	})
	assert.Equal(t, 2, len(h.components))

	expected := []Component{
		{
			Name:     "redis",
			Critical: false,
			Check: func(ctx context.Context) Status {
				return StatusUp
			},
		},
		{
			Name:     "mongo",
			Critical: true,
			Check: func(ctx context.Context) Status {
				return StatusUp
			},
		},
	}

	for i := range h.components {
		assert.Equal(t, expected[i].Name, h.components[i].Name)
		assert.Equal(t, expected[i].Critical, h.components[i].Critical)
	}
}

func TestHealth_Status(t *testing.T) {
	tests := []struct {
		name     string
		init     func() *Health
		expected Status
	}{
		{
			name: "All Up",
			init: func() *Health {
				return New(
					Component{
						Name:     "redis",
						Critical: false,
						Check: func(ctx context.Context) Status {
							return StatusUp
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) Status {
							return StatusUp
						},
					},
				)
			},
			expected: StatusUp,
		},
		{
			name: "Critical Component Degraded",
			init: func() *Health {
				return New(
					Component{
						Name:     "redis",
						Critical: false,
						Check: func(ctx context.Context) Status {
							return StatusUp
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) Status {
							return StatusDegraded
						},
					},
				)
			},
			expected: StatusDegraded,
		},
		{
			name: "Critical Component Down",
			init: func() *Health {
				return New(
					Component{
						Name:     "redis",
						Critical: false,
						Check: func(ctx context.Context) Status {
							return StatusUp
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) Status {
							return StatusDown
						},
					},
				)
			},
			expected: StatusDown,
		},
		{
			name: "Non Critical Component Down",
			init: func() *Health {
				return New(
					Component{
						Name:     "redis",
						Critical: false,
						Check: func(ctx context.Context) Status {
							return StatusDown
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) Status {
							return StatusUp
						},
					},
				)
			},
			expected: StatusDegraded,
		},
		{
			name: "All Down",
			init: func() *Health {
				return New(
					Component{
						Name:     "redis",
						Critical: false,
						Check: func(ctx context.Context) Status {
							return StatusDown
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) Status {
							return StatusDown
						},
					},
				)
			},
			expected: StatusDown,
		},
		{
			name: "Non Critical Component Degraded",
			init: func() *Health {
				return New(
					Component{
						Name:     "redis",
						Critical: false,
						Check: func(ctx context.Context) Status {
							return StatusDegraded
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) Status {
							return StatusUp
						},
					},
				)
			},
			expected: StatusDegraded,
		},
		{
			name: "Critical Down and Non Critical Degraded",
			init: func() *Health {
				return New(
					Component{
						Name:     "redis",
						Critical: false,
						Check: func(ctx context.Context) Status {
							return StatusDegraded
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) Status {
							return StatusDown
						},
					},
				)
			},
			expected: StatusDown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.init()
			actual := h.Status(context.Background())
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestHealth_ServeHTTP(t *testing.T) {
	type response struct {
		Status     Status            `json:"status"`
		Uptime     string            `json:"uptime"`
		Components []ComponentStatus `json:"components"`
	}

	tests := []struct {
		name               string
		init               func() *Health
		expectedHttpStatus int
		expectedResponse   response
	}{
		{
			name: "All Up",
			init: func() *Health {
				return New(
					Component{
						Name:     "redis",
						Critical: false,
						Check: func(ctx context.Context) Status {
							return StatusUp
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) Status {
							return StatusUp
						},
					},
				)
			},
			expectedHttpStatus: http.StatusOK,
			expectedResponse: response{
				Status: StatusUp,
				Components: []ComponentStatus{
					{
						Name:     "redis",
						Critical: false,
						Status:   StatusUp,
					},
					{
						Name:     "mongo",
						Critical: true,
						Status:   StatusUp,
					},
				},
			},
		},
		{
			name: "Critical Component Down",
			init: func() *Health {
				return New(
					Component{
						Name:     "redis",
						Critical: false,
						Check: func(ctx context.Context) Status {
							return StatusUp
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) Status {
							return StatusDown
						},
					},
				)
			},
			expectedHttpStatus: http.StatusServiceUnavailable,
			expectedResponse: response{
				Status: StatusDown,
				Components: []ComponentStatus{
					{
						Name:     "redis",
						Critical: false,
						Status:   StatusUp,
					},
					{
						Name:     "mongo",
						Critical: true,
						Status:   StatusDown,
					},
				},
			},
		},
		{
			name: "Non Critical Component Down",
			init: func() *Health {
				return New(
					Component{
						Name:     "redis",
						Critical: false,
						Check: func(ctx context.Context) Status {
							return StatusUp
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) Status {
							return StatusDown
						},
					},
				)
			},
			expectedHttpStatus: http.StatusServiceUnavailable,
			expectedResponse: response{
				Status: StatusDown,
				Components: []ComponentStatus{
					{
						Name:     "redis",
						Critical: false,
						Status:   StatusUp,
					},
					{
						Name:     "mongo",
						Critical: true,
						Status:   StatusDown,
					},
				},
			},
		},
		{
			name: "All Components Down",
			init: func() *Health {
				return New(
					Component{
						Name:     "redis",
						Critical: false,
						Check: func(ctx context.Context) Status {
							return StatusDown
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) Status {
							return StatusDown
						},
					},
				)
			},
			expectedHttpStatus: http.StatusServiceUnavailable,
			expectedResponse: response{
				Status: StatusDown,
				Components: []ComponentStatus{
					{
						Name:     "redis",
						Critical: false,
						Status:   StatusDown,
					},
					{
						Name:     "mongo",
						Critical: true,
						Status:   StatusDown,
					},
				},
			},
		},
		{
			name: "Non Critical Component Degraded",
			init: func() *Health {
				return New(
					Component{
						Name:     "redis",
						Critical: false,
						Check: func(ctx context.Context) Status {
							return StatusDegraded
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) Status {
							return StatusUp
						},
					},
				)
			},
			expectedHttpStatus: http.StatusOK,
			expectedResponse: response{
				Status: StatusDegraded,
				Components: []ComponentStatus{
					{
						Name:     "redis",
						Critical: false,
						Status:   StatusDegraded,
					},
					{
						Name:     "mongo",
						Critical: true,
						Status:   StatusUp,
					},
				},
			},
		},
		{
			name: "Critical Component Degraded",
			init: func() *Health {
				return New(
					Component{
						Name:     "redis",
						Critical: false,
						Check: func(ctx context.Context) Status {
							return StatusUp
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) Status {
							return StatusDegraded
						},
					},
				)
			},
			expectedHttpStatus: http.StatusOK,
			expectedResponse: response{
				Status: StatusDegraded,
				Components: []ComponentStatus{
					{
						Name:     "redis",
						Critical: false,
						Status:   StatusUp,
					},
					{
						Name:     "mongo",
						Critical: true,
						Status:   StatusDegraded,
					},
				},
			},
		},
		{
			name: "All Degraded",
			init: func() *Health {
				return New(
					Component{
						Name:     "redis",
						Critical: false,
						Check: func(ctx context.Context) Status {
							return StatusDegraded
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) Status {
							return StatusDegraded
						},
					},
				)
			},
			expectedHttpStatus: http.StatusOK,
			expectedResponse: response{
				Status: StatusDegraded,
				Components: []ComponentStatus{
					{
						Name:     "redis",
						Critical: false,
						Status:   StatusDegraded,
					},
					{
						Name:     "mongo",
						Critical: true,
						Status:   StatusDegraded,
					},
				},
			},
		},
		{
			name: "One Critical Up One Critical Degraded",
			init: func() *Health {
				return New(
					Component{
						Name:     "redis",
						Critical: true,
						Check: func(ctx context.Context) Status {
							return StatusDown
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) Status {
							return StatusUp
						},
					},
				)
			},
			expectedHttpStatus: http.StatusServiceUnavailable,
			expectedResponse: response{
				Status: StatusDown,
				Components: []ComponentStatus{
					{
						Name:     "redis",
						Critical: true,
						Status:   StatusDown,
					},
					{
						Name:     "mongo",
						Critical: true,
						Status:   StatusUp,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.init()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			h.ServeHTTP(w, r)
			assert.Equal(t, tt.expectedHttpStatus, w.Code)

			var res response
			err := json.Unmarshal(w.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResponse.Status, res.Status)
			assert.Equal(t, tt.expectedResponse.Components, res.Components)
		})
	}
}
