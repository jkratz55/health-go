package health

import (
	"context"
)

// A Check is a type that provides its health status.
//
// Implementations of Check should return either StatusUp, StatusDegraded, or
// StatusDown depending on the health of the component.
type Check interface {
	Status(ctx context.Context) Status
}

// CheckFunc is an adapter to allow the use of ordinary functions as an
// implementation of the Check interface.
type CheckFunc func(ctx context.Context) Status

func (cf CheckFunc) Status(ctx context.Context) Status {
	return cf(ctx)
}
