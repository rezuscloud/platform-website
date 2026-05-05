package obs

import "context"

// Client fetches live service map data from SigNoz metrics.
type Client interface {
	Fetch(ctx context.Context) (LiveData, error)
}
