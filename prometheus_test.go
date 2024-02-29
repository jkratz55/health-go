package health

import (
	"context"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

type mockComponent struct {
	status Status
}

func (m *mockComponent) Status(ctx context.Context) Status {
	return m.status
}

func TestEnablePrometheus(t *testing.T) {

	redisCheck := &mockComponent{
		status: StatusUp,
	}
	mongoCheck := &mockComponent{
		status: StatusUp,
	}

	hc := New()

	hc.Register(Component{
		Name:     "redis",
		Critical: false,
		Check:    redisCheck.Status,
	})
	hc.Register(Component{
		Name:     "mongo",
		Critical: true,
		Check:    mongoCheck.Status,
	})

	err := EnablePrometheus(hc)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		init     func()
		expected string
	}{
		{
			name: "All Components Up",
			init: func() {
				redisCheck.status = StatusUp
				mongoCheck.status = StatusUp
			},
			expected: `
				# HELP health_status Indicator of overall status of the application instance. 0 is down, 1 is degraded, 2 is up.
				# TYPE health_status gauge
				health_status 2
			`,
		},
		{
			name: "All Components Down",
			init: func() {
				redisCheck.status = StatusDown
				mongoCheck.status = StatusDown
			},
			expected: `
				# HELP health_status Indicator of overall status of the application instance. 0 is down, 1 is degraded, 2 is up.
				# TYPE health_status gauge
				health_status 0
			`,
		},
		{
			name: "Non Critical Component Down",
			init: func() {
				redisCheck.status = StatusDown
				mongoCheck.status = StatusUp
			},
			expected: `
				# HELP health_status Indicator of overall status of the application instance. 0 is down, 1 is degraded, 2 is up.
				# TYPE health_status gauge
				health_status 1
			`,
		},
		{
			name: "Critical Component Degraded",
			init: func() {
				redisCheck.status = StatusUp
				mongoCheck.status = StatusDegraded

			},
			expected: `
				# HELP health_status Indicator of overall status of the application instance. 0 is down, 1 is degraded, 2 is up.
				# TYPE health_status gauge
				health_status 1
			`,
		},
		{
			name: "Non Critical Component Degraded",
			init: func() {
				redisCheck.status = StatusDegraded
				mongoCheck.status = StatusUp

			},
			expected: `
				# HELP health_status Indicator of overall status of the application instance. 0 is down, 1 is degraded, 2 is up.
				# TYPE health_status gauge
				health_status 1
			`,
		},
		{
			name: "All Components Degraded",
			init: func() {
				redisCheck.status = StatusDegraded
				mongoCheck.status = StatusDegraded
			},
			expected: `
				# HELP health_status Indicator of overall status of the application instance. 0 is down, 1 is degraded, 2 is up.
				# TYPE health_status gauge
				health_status 1
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.init()
			err := testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(tt.expected), "health_status")
			assert.NoError(t, err)
		})
	}

	// name := []struct {
	// 	name     string
	// 	init     func() *Health
	// 	expected string
	// }{
	// 	{
	// 		name: "All Components Up",
	// 		init: func() *Health {
	// 			hc := New()
	// 			hc.Register(Component{
	// 				Name:     "redis",
	// 				Critical: true,
	// 				Check:    redisCheck.Status,
	// 			})
	// 			hc.Register(Component{
	// 				Name:     "mongo",
	// 				Critical: true,
	// 				Check:    mongoCheck.Status,
	// 			})
	// 			err := EnablePrometheus(hc)
	// 			assert.NoError(t, err)
	// 			return hc
	// 		},
	// 		expected: `
	// 			# HELP health_status Indicator of overall status of the application instance. 0 is down, 1 is degraded, 2 is up.
	// 			# TYPE health_status gauge
	// 			health_status 2
	// 		`,
	// 	},
	// 	{
	// 		name: "All Components Down",
	// 		init: func() *Health {
	// 			hc := New()
	// 			hc.Register(Component{
	// 				Name:     "redis",
	// 				Critical: true,
	// 				Check: func(ctx context.Context) Status {
	// 					return StatusDown
	// 				},
	// 			})
	// 			hc.Register(Component{
	// 				Name:     "mongo",
	// 				Critical: true,
	// 				Check: func(ctx context.Context) Status {
	// 					return StatusDown
	// 				},
	// 			})
	// 			err := EnablePrometheus(hc)
	// 			assert.NoError(t, err)
	// 			return hc
	// 		},
	// 		expected: `
	// 			# HELP health_status Indicator of overall status of the application instance. 0 is down, 1 is degraded, 2 is up.
	// 			# TYPE health_status gauge
	// 			health_status 0 lkjhgfdesw		123
	// 		`,
	// 	},
	// }
	//
	// for _, tt := range name {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		_ = tt.init()
	// 		err := testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(tt.expected), "health_status")
	// 		assert.NoError(t, err)
	// 	})
	//
	// }

	// h := New()
	// err := EnablePrometheus(h)
	// assert.NoError(t, err)

	// promHandler := promhttp.Handler()
	// w := httptest.NewRecorder()
	// r := httptest.NewRequest(http.MethodGet, "/", nil)
	// promHandler.ServeHTTP(w, r)
	// assert.Equal(t, http.StatusOK, w.Code)
	//
	// err = testutil.GatherAndCompare(prometheus.DefaultGatherer, w.Body, "health")
	// assert.NoError(t, err)

	// expected := `
	//     # HELP health_status Indicator of overall status of the application instance. 0 is down, 1 is degraded, 2 is up.
	//     # TYPE health_status gauge
	//     health_status 2
	// `

	// err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(expected), "health_status")
	// assert.NoError(t, err)
}
