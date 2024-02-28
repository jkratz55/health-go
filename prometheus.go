package health

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

// EnablePrometheus exposes the overall status and status of each component as
// Prometheus metrics using the default Prometheus registry.
//
// The overall status is exposed as a gauge named "health_status" with a value
// of 0 for down, 1 for degraded, and 2 for up.
//
// The status of each component is exposed as a gauge named "health_component_status"
// with a value of 0 for down, 1 for degraded, and 2 for up.
func EnablePrometheus(h *Health) error {
	overallStatus := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "health",
		Name:      "status",
		Help:      "Indicator of overall status of the application instance. 0 is down, 1 is degraded, 2 is up.",
	})
	componentStatus := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "health",
		Name:      "component_status",
		Help:      "Indicator of status of the application components. 0 is down, 1 is degraded, 2 is up.",
	}, []string{"component"})

	c := &collector{
		components: h.components,
		overall:    overallStatus,
		component:  componentStatus,
	}
	return prometheus.Register(c)
}

type collector struct {
	components Components
	overall    prometheus.Gauge
	component  *prometheus.GaugeVec
}

func (c collector) Describe(descs chan<- *prometheus.Desc) {
	c.overall.Describe(descs)
	c.component.Describe(descs)
}

func (c collector) Collect(metrics chan<- prometheus.Metric) {
	overall := c.components.Status(context.Background())
	switch overall {
	case StatusDown:
		c.overall.Set(0)
	case StatusDegraded:
		c.overall.Set(1)
	case StatusUp:
		c.overall.Set(2)
	}

	componentStatuses := c.components.ComponentStatus(context.Background())
	for _, status := range componentStatuses {
		switch status.Status {
		case StatusDown:
			c.component.WithLabelValues(status.Name).Set(0)
		case StatusDegraded:
			c.component.WithLabelValues(status.Name).Set(1)
		case StatusUp:
			c.component.WithLabelValues(status.Name).Set(2)
		}
	}

	c.overall.Collect(metrics)
	c.component.Collect(metrics)
}
