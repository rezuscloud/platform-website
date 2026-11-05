package obs

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// newTestTopology builds a k8sTopology whose API calls hit srv, seeded with an
// existing pod map and an already-expired cache.
func newTestTopology(t *testing.T, srv *httptest.Server, seedPods map[string]PodInfo) *k8sTopology {
	t.Helper()
	kt := &k8sTopology{
		topo:    ClusterTopology{Nodes: map[string]NodeInfo{}, Pods: seedPods},
		expires: time.Now().Add(-time.Minute), // stale
		ttl:     5 * time.Minute,
		baseURL: srv.URL,
		client:  srv.Client(),
	}
	return kt
}

func testNodesPayload(nodes int) string {
	body := `{"items":[`
	for i := 0; i < nodes; i++ {
		if i > 0 {
			body += ","
		}
		body += fmt.Sprintf(`{"metadata":{"name":"node-%d","creationTimestamp":"2026-01-01T00:00:00Z"},`+
			`"spec":{"providerID":""},"status":{"nodeInfo":{"architecture":"amd64"},`+
			`"capacity":{"cpu":"8","memory":"32Gi"}}}`, i)
	}
	return body + "]}"
}

func testPodsPayload(pods int) string {
	body := `{"items":[`
	for i := 0; i < pods; i++ {
		if i > 0 {
			body += ","
		}
		body += fmt.Sprintf(`{"metadata":{"name":"pod-%d","namespace":"default",`+
			`"ownerReferences":[{"kind":"ReplicaSet","name":"rs-%d","uid":"u"}]},`+
			`"spec":{"nodeName":"node-0"}}`, i, i)
	}
	return body + "]}"
}

// A failed pod fetch must keep the previously known pods and schedule a short
// retry — never cache an empty pod map for the full TTL.
func TestRefreshKeepsStalePodsOnFailure(t *testing.T) {
	var podHits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/nodes":
			w.Write([]byte(testNodesPayload(1)))
		case "/api/v1/pods":
			atomic.AddInt32(&podHits, 1)
			w.WriteHeader(http.StatusTooManyRequests) // apiserver APF throttle
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	seed := map[string]PodInfo{"default/old": {Deployment: "old-deploy", Node: "node-0"}}
	kt := newTestTopology(t, srv, seed)

	if err := kt.refresh(context.Background()); err != nil {
		t.Fatalf("refresh should not fail when only the pod fetch fails: %v", err)
	}

	kt.mu.RLock()
	gotPods, expires := len(kt.topo.Pods), kt.expires
	kt.mu.RUnlock()

	if gotPods != 1 {
		t.Fatalf("seeded pods must survive a failed refresh, got %d pods", gotPods)
	}
	if d := time.Until(expires); d > 35*time.Second {
		t.Fatalf("failed refresh must retry soon, expires in %s", d)
	}
	if atomic.LoadInt32(&podHits) == 0 {
		t.Fatal("pod endpoint was never called")
	}
}

// Concurrent Topology() callers must not fan out into concurrent refreshes.
func TestTopologySingleflight(t *testing.T) {
	var nodeHits int32
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/nodes":
			atomic.AddInt32(&nodeHits, 1)
			<-release // hold each refresh open
			w.Write([]byte(testNodesPayload(1)))
		case "/api/v1/pods":
			w.Write([]byte(testPodsPayload(2)))
		case "/apis/apps/v1/replicasets":
			w.Write([]byte(`{"items":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	kt := newTestTopology(t, srv, map[string]PodInfo{})

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			kt.Topology()
		}()
	}
	// give the goroutines a moment to pile onto the single in-flight refresh
	time.Sleep(100 * time.Millisecond)
	close(release)
	wg.Wait()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && kt.refreshing.Load() {
		time.Sleep(10 * time.Millisecond)
	}
	if hits := atomic.LoadInt32(&nodeHits); hits > 2 {
		t.Fatalf("expected at most one (retry-tolerant) refresh, node endpoint hit %d times", hits)
	}
}

// A successful refresh replaces the pod map and honors the full TTL.
func TestRefreshSuccessUsesTTL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/nodes":
			w.Write([]byte(testNodesPayload(1)))
		case "/api/v1/pods":
			w.Write([]byte(testPodsPayload(2)))
		case "/apis/apps/v1/replicasets":
			w.Write([]byte(`{"items":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	kt := newTestTopology(t, srv, map[string]PodInfo{})
	if err := kt.refresh(context.Background()); err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	kt.mu.RLock()
	pods, expires := len(kt.topo.Pods), kt.expires
	kt.mu.RUnlock()

	if pods != 2 {
		t.Fatalf("expected 2 pods, got %d", pods)
	}
	if d := time.Until(expires); d < 4*time.Minute {
		t.Fatalf("successful refresh must honor the TTL, expires in %s", d)
	}
}
