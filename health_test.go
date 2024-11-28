package health

import (
	"context"
	"encoding/json"
	"errors"
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
					Check: func(ctx context.Context) error {
						return nil
					},
				})
				components = append(components, Component{
					Name:     "mongo",
					Critical: true,
					Check: func(ctx context.Context) error {
						return nil
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
		Check: func(ctx context.Context) error {
			return nil
		},
	})
	h.Register(Component{
		Name:     "mongo",
		Critical: true,
		Check: func(ctx context.Context) error {
			return nil
		},
	})
	assert.Equal(t, 2, len(h.components))

	expected := []Component{
		{
			Name:     "redis",
			Critical: false,
			Check: func(ctx context.Context) error {
				return nil
			},
		},
		{
			Name:     "mongo",
			Critical: true,
			Check: func(ctx context.Context) error {
				return nil
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
						Check: func(ctx context.Context) error {
							return nil
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) error {
							return nil
						},
					},
				)
			},
			expected: StatusUp,
		},
		{
			name: "Critical Component Down",
			init: func() *Health {
				return New(
					Component{
						Name:     "redis",
						Critical: false,
						Check: func(ctx context.Context) error {
							return nil
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) error {
							return errors.New("mongo down")
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
						Check: func(ctx context.Context) error {
							return errors.New("redis down")
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) error {
							return nil
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
						Check: func(ctx context.Context) error {
							return errors.New("redis down")
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) error {
							return errors.New("mongo down")
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
			for i := 0; i < len(h.components); i++ {
				err := h.components[i].Check(context.Background())
				if err == nil {
					h.components[i].status = StatusUp
				} else {
					h.components[i].status = StatusDown
				}
			}
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
						Check: func(ctx context.Context) error {
							return nil
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) error {
							return nil
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
						Check: func(ctx context.Context) error {
							return nil
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) error {
							return errors.New("mongo down")
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
						Check: func(ctx context.Context) error {
							return errors.New("redis down")
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) error {
							return nil
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
		{
			name: "All Components Down",
			init: func() *Health {
				return New(
					Component{
						Name:     "redis",
						Critical: false,
						Check: func(ctx context.Context) error {
							return errors.New("redis down")
						},
					},
					Component{
						Name:     "mongo",
						Critical: true,
						Check: func(ctx context.Context) error {
							return errors.New("mongo down")
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.init()
			for i := 0; i < len(h.components); i++ {
				err := h.components[i].Check(context.Background())
				if err == nil {
					h.components[i].status = StatusUp
				} else {
					h.components[i].status = StatusDown
				}
			}
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
