package obs

import "context"

// Client fetches live infrastructure data from SigNoz metrics.
type Client interface {
	// Fetch returns the current cluster topology and metrics.
	Fetch(ctx context.Context) (LiveData, error)
}
