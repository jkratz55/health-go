package health

import (
	"context"
)

// CheckFunc is a function type that provides the health status of a component.
//
// A CheckFunc should return either StatusUp, StatusDegraded, or StatusDown
// depending on the health of the component.
type CheckFunc func(ctx context.Context) Status
