package health

import (
	"context"
)

// CheckFunc is a function type that checks/verifies the health of a component
// and or service. If an error is returned, the component/service is considered
// unhealthy and down.
type CheckFunc func(ctx context.Context) error
