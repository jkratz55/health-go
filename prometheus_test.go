package health

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

type mockComponent struct {
	err error
}

func (m *mockComponent) Status(ctx context.Context) error {
	return m.err
}

func TestEnablePrometheus(t *testing.T) {

	redisCheck := &mockComponent{
		err: nil,
	}
	mongoCheck := &mockComponent{
		err: nil,
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
				redisCheck.err = nil
				mongoCheck.err = nil
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
				redisCheck.err = errors.New("redis down")
				mongoCheck.err = errors.New("mongo down")
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
				redisCheck.err = errors.New("redis down")
				mongoCheck.err = nil
			},
			expected: `
				# HELP health_status Indicator of overall status of the application instance. 0 is down, 1 is degraded, 2 is up.
				# TYPE health_status gauge
				health_status 1
			`,
		},
		{
			name: "Critical Component Down",
			init: func() {
				redisCheck.err = nil
				mongoCheck.err = errors.New("mongo down")
			},
			expected: `
				# HELP health_status Indicator of overall status of the application instance. 0 is down, 1 is degraded, 2 is up.
				# TYPE health_status gauge
				health_status 0
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.init()
			for i := 0; i < len(hc.components); i++ {
				err := hc.components[i].Check(context.Background())
				if err == nil {
					hc.components[i].status = StatusUp
				} else {
					hc.components[i].status = StatusDown
				}
			}
			err := testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(tt.expected), "health_status")
			assert.NoError(t, err)
		})
	}
}
