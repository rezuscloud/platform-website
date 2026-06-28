package obs

import (
	"context"
	"time"
)

// MockClient returns the platform topology without live data.
type MockClient struct {
	Data LiveData
	Err  error
}

func (m *MockClient) Fetch(_ context.Context) (LiveData, error) {
	return m.Data, m.Err
}

// DefaultMockData returns an empty snapshot used as the fallback when the live
// topology source is unavailable (empty result or fetch error).
//
// It deliberately returns NO hosts and NO services. This site's central claim
// (stated on /privacy) is that the live grid shows our real infrastructure.
// Fabricating a plausible-looking cluster on failure would silently break that
// promise, so the fallback surfaces "topology unavailable" via the section's
// degraded rendering (HasMetrics=false) instead of inventing nodes.
//
// Dev/preview environments that want a populated sample should inject their
// own Client via SetLiveClient, not rely on this fallback.
func DefaultMockData() LiveData {
	return LiveData{
		SelfNamespace: "platform-website",
		HasMetrics:    false,
		Timestamp:     time.Now().Unix(),
	}
}
