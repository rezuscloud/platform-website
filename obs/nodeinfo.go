package obs

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// k8sNodeInfo implements NodeInfoFunc by querying the Kubernetes API.
// It caches results and refreshes periodically.
type k8sNodeInfo struct {
	mu      sync.RWMutex
	nodes   map[string]NodeInfo // node name -> info
	expires time.Time
	ttl     time.Duration
	baseURL string
	token   string
	client  *http.Client
}

// NewK8sNodeInfo creates a NodeInfoFunc that queries the in-cluster
// Kubernetes API for node roles. Returns nil if not running in a cluster.
func NewK8sNodeInfo() NodeInfoFunc {
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	if host == "" {
		log.Println("nodeinfo: not in cluster, using hostname-based node role detection")
		return nil
	}

	tokenBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		log.Printf("nodeinfo: cannot read service account token: %v", err)
		return nil
	}

	tlsCfg, err := tlsConfigFromCA()
	if err != nil {
		log.Printf("nodeinfo: cannot build TLS config: %v", err)
		return nil
	}

	ki := &k8sNodeInfo{
		nodes:   make(map[string]NodeInfo),
		ttl:     5 * time.Minute,
		baseURL: fmt.Sprintf("https://%s:%s", host, os.Getenv("KUBERNETES_SERVICE_PORT")),
		token:   strings.TrimSpace(string(tokenBytes)),
		client: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: tlsCfg,
			},
		},
	}

	// Initial fetch
	if err := ki.refresh(context.Background()); err != nil {
		log.Printf("nodeinfo: initial fetch failed: %v", err)
	}

	return ki.lookup
}

func (ki *k8sNodeInfo) lookup(name string) (NodeInfo, bool) {
	ki.mu.RLock()
	defer ki.mu.RUnlock()

	if time.Now().After(ki.expires) {
		// Trigger async refresh
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := ki.refresh(ctx); err != nil {
				log.Printf("nodeinfo: refresh failed: %v", err)
			}
		}()
	}

	info, ok := ki.nodes[name]
	return info, ok
}

func (ki *k8sNodeInfo) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", ki.baseURL+"/api/v1/nodes", nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+ki.token)

	resp, err := ki.client.Do(req)
	if err != nil {
		return fmt.Errorf("query nodes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result struct {
		Items []struct {
			Metadata struct {
				Name   string            `json:"name"`
				Labels map[string]string `json:"labels"`
			} `json:"metadata"`
			Spec struct {
				ProviderID string `json:"providerID"`
			} `json:"spec"`
			Status struct {
				NodeInfo struct {
					Architecture string `json:"architecture"`
				} `json:"nodeInfo"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	nodes := make(map[string]NodeInfo, len(result.Items))
	for _, item := range result.Items {
		_, isCP := item.Metadata.Labels["node-role.kubernetes.io/control-plane"]
		if !isCP {
			_, isCP = item.Metadata.Labels["node-role.kubernetes.io/master"]
		}

		arch := "AMD64"
		if strings.Contains(strings.ToLower(item.Status.NodeInfo.Architecture), "arm") {
			arch = "ARM64"
		}

		provider := providerFromID(item.Spec.ProviderID)
		if provider == "" {
			provider = providerFromName(item.Metadata.Name)
		}

		nodes[item.Metadata.Name] = NodeInfo{
			IsControlPlane: isCP,
			Provider:       provider,
			Arch:           arch,
		}
	}

	ki.mu.Lock()
	ki.nodes = nodes
	ki.expires = time.Now().Add(ki.ttl)
	ki.mu.Unlock()

	log.Printf("nodeinfo: refreshed %d nodes", len(nodes))
	return nil
}

// providerFromID extracts the hosting provider from the Kubernetes
// spec.providerID field. Examples:
//   - "ocid1.instance.oc1..." -> "OCI Cloud"
//   - "metal://edge"         -> "Edge"
//   - "proxmox://..."        -> "Proxmox"
func providerFromID(id string) string {
	switch {
	case strings.HasPrefix(id, "ocid1."):
		return "OCI Cloud"
	case strings.HasPrefix(id, "metal://"):
		return "Edge"
	case strings.Contains(id, "proxmox"):
		return "Proxmox"
	default:
		return ""
	}
}

// providerFromName guesses the hosting provider from the Talos machine name.
// Talos OCI machines have names like "talos-oci-c-*" or "talosoci-*".
// Talos edge machines have names like "talosedge-*".
// Proxmox VMs have names like "talos-os-w-*" or "talos-amd64-*".
// This is a fallback when providerID is not available.
func providerFromName(name string) string {
	switch {
	case strings.HasPrefix(name, "talos-oci") || strings.HasPrefix(name, "talosoci"):
		return "OCI Cloud"
	case strings.HasPrefix(name, "talosedge"):
		return "Edge"
	case strings.HasPrefix(name, "talos-os-") || strings.HasPrefix(name, "talos-amd64"):
		return "Cloud"
	default:
		return "Node"
	}
}

// tlsConfigFromCA builds a TLS config that trusts the in-cluster CA.
// Returns an error (instead of falling back to InsecureSkipVerify) if
// the CA cert cannot be read.
func tlsConfigFromCA() (*tls.Config, error) {
	caBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		return nil, fmt.Errorf("read CA cert: %w", err)
	}
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(caBytes) {
		return nil, fmt.Errorf("no certificates found in CA bundle")
	}
	return &tls.Config{RootCAs: cp}, nil
}
