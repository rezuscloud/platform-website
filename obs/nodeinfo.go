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
	"strconv"
	"strings"
	"sync"
	"time"
)

// ownerRef is a Kubernetes metadata.ownerReference entry.
type ownerRef struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

// ClusterTopology is the native view of the cluster from the Kubernetes API:
// node metadata (role, provider, arch, capacity, age) and pod ownership
// (which deployment a pod belongs to, which node it runs on).
// No metrics here — only topology and capacity.
type ClusterTopology struct {
	Nodes map[string]NodeInfo // node name -> metadata + capacity
	Pods  map[string]PodInfo  // "namespace/pod" -> deployment + node
}

// PodInfo maps a running pod to its owning workload and node.
type PodInfo struct {
	Deployment string // deployment / statefulset / daemonset name
	Node       string
}

// k8sTopology queries the in-cluster Kubernetes API for node metadata and pod
// ownership. Results are cached and refreshed periodically.
type k8sTopology struct {
	mu      sync.RWMutex
	topo    ClusterTopology
	expires time.Time
	ttl     time.Duration
	baseURL string
	token   string
	client  *http.Client
}

// NewK8sTopology creates a topology provider that queries the in-cluster
// Kubernetes API. Returns nil if not running in a cluster (dev / local).
func NewK8sTopology() *k8sTopology {
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	if host == "" {
		log.Println("topology: not in cluster, using mock topology")
		return nil
	}

	tokenBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		log.Printf("topology: cannot read service account token: %v", err)
		return nil
	}

	tlsCfg, err := tlsConfigFromCA()
	if err != nil {
		log.Printf("topology: cannot build TLS config: %v", err)
		return nil
	}

	kt := &k8sTopology{
		topo:    ClusterTopology{Nodes: map[string]NodeInfo{}, Pods: map[string]PodInfo{}},
		ttl:     5 * time.Minute,
		baseURL: fmt.Sprintf("https://%s:%s", host, os.Getenv("KUBERNETES_SERVICE_PORT")),
		token:   strings.TrimSpace(string(tokenBytes)),
		client: &http.Client{
			Timeout: 8 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: tlsCfg,
			},
		},
	}

	if err := kt.refresh(context.Background()); err != nil {
		log.Printf("topology: initial fetch failed: %v", err)
	}

	return kt
}

// Topology returns the cached cluster topology, refreshing async when stale.
func (kt *k8sTopology) Topology() ClusterTopology {
	kt.mu.RLock()
	defer kt.mu.RUnlock()

	if time.Now().After(kt.expires) {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			if err := kt.refresh(ctx); err != nil {
				log.Printf("topology: refresh failed: %v", err)
			}
		}()
	}
	return kt.topo
}

// NodeInfoFunc adapts the topology into the lookup signature used by
// BuildHosts. Returns (NodeInfo, true) for known nodes.
func (kt *k8sTopology) NodeInfoFunc() NodeInfoFunc {
	return func(name string) (NodeInfo, bool) {
		kt.mu.RLock()
		defer kt.mu.RUnlock()
		info, ok := kt.topo.Nodes[name]
		return info, ok
	}
}

func (kt *k8sTopology) refresh(ctx context.Context) error {
	nodes, err := kt.fetchNodes(ctx)
	if err != nil {
		return fmt.Errorf("fetch nodes: %w", err)
	}
	pods, err := kt.fetchPodTopology(ctx)
	if err != nil {
		// Topology without pod ownership still degrades gracefully.
		log.Printf("topology: fetch pod ownership failed: %v", err)
		pods = map[string]PodInfo{}
	}

	kt.mu.Lock()
	kt.topo = ClusterTopology{Nodes: nodes, Pods: pods}
	kt.expires = time.Now().Add(kt.ttl)
	kt.mu.Unlock()

	log.Printf("topology: refreshed %d nodes, %d pods", len(nodes), len(pods))
	return nil
}

func (kt *k8sTopology) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", kt.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+kt.token)
	req.Header.Set("Accept", "application/json")

	resp, err := kt.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d for %s", resp.StatusCode, path)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (kt *k8sTopology) fetchNodes(ctx context.Context) (map[string]NodeInfo, error) {
	var result struct {
		Items []struct {
			Metadata struct {
				Name              string            `json:"name"`
				Labels            map[string]string `json:"labels"`
				CreationTimestamp string            `json:"creationTimestamp"`
			} `json:"metadata"`
			Spec struct {
				ProviderID string `json:"providerID"`
			} `json:"spec"`
			Status struct {
				NodeInfo struct {
					Architecture string `json:"architecture"`
				} `json:"nodeInfo"`
				Capacity map[string]string `json:"capacity"`
			} `json:"status"`
		} `json:"items"`
	}
	if err := kt.get(ctx, "/api/v1/nodes", &result); err != nil {
		return nil, err
	}

	nodes := make(map[string]NodeInfo, len(result.Items))
	for _, item := range result.Items {
		_, isCP := item.Metadata.Labels["node-role.kubernetes.io/control-plane"]
		if !isCP {
			_, isCP = item.Metadata.Labels["node-role.kubernetes.io/master"]
		}

		nodes[item.Metadata.Name] = NodeInfo{
			IsControlPlane: isCP,
			Provider:       providerFromID(item.Spec.ProviderID, item.Metadata.Name),
			Arch:           archLabel(item.Status.NodeInfo.Architecture),
			CPUCores:       parseCores(item.Status.Capacity["cpu"]),
			RAMBytes:       parseBytes(item.Status.Capacity["memory"]),
			Created:        parseTime(item.Metadata.CreationTimestamp),
		}
	}
	return nodes, nil
}

// fetchPodTopology resolves pod -> owning workload + node.
// It queries pods and replicasets so a ReplicaSet-owned pod maps back to its
// Deployment (pods own ReplicaSets, ReplicaSets own Deployments).
func (kt *k8sTopology) fetchPodTopology(ctx context.Context) (map[string]PodInfo, error) {
	var podList struct {
		Items []struct {
			Metadata struct {
				Name      string     `json:"name"`
				Namespace string     `json:"namespace"`
				OwnerRefs []ownerRef `json:"ownerReferences"`
			} `json:"metadata"`
			Spec struct {
				NodeName string `json:"nodeName"`
			} `json:"spec"`
		} `json:"items"`
	}
	if err := kt.get(ctx, "/api/v1/pods", &podList); err != nil {
		return nil, err
	}

	rsToDeployment := map[string]string{} // "namespace/rs" -> deployment
	if rs, err := kt.fetchReplicaSetOwners(ctx); err == nil {
		rsToDeployment = rs
	}

	pods := make(map[string]PodInfo, len(podList.Items))
	for _, p := range podList.Items {
		if p.Spec.NodeName == "" {
			continue // pending / unscheduled
		}
		pods[p.Metadata.Namespace+"/"+p.Metadata.Name] = PodInfo{
			Deployment: resolveWorkload(p.Metadata.OwnerRefs, p.Metadata.Namespace, rsToDeployment),
			Node:       p.Spec.NodeName,
		}
	}
	return pods, nil
}

func (kt *k8sTopology) fetchReplicaSetOwners(ctx context.Context) (map[string]string, error) {
	var rsList struct {
		Items []struct {
			Metadata struct {
				Name      string     `json:"name"`
				Namespace string     `json:"namespace"`
				OwnerRefs []ownerRef `json:"ownerReferences"`
			} `json:"metadata"`
		} `json:"items"`
	}
	if err := kt.get(ctx, "/apis/apps/v1/replicasets", &rsList); err != nil {
		return nil, err
	}
	m := make(map[string]string, len(rsList.Items))
	for _, rs := range rsList.Items {
		for _, o := range rs.Metadata.OwnerRefs {
			if o.Kind == "Deployment" {
				m[rs.Metadata.Namespace+"/"+rs.Metadata.Name] = o.Name
				break
			}
		}
	}
	return m, nil
}

// resolveWorkload turns a pod's ownerReferences into a workload name.
//   - Deployment: pod -> ReplicaSet -> Deployment (via rsToDeployment)
//   - StatefulSet / DaemonSet / Job: owner name is the workload itself
//   - Bare pod: empty
func resolveWorkload(refs []ownerRef, namespace string, rsToDeployment map[string]string) string {
	for _, o := range refs {
		switch o.Kind {
		case "ReplicaSet":
			if dep, ok := rsToDeployment[namespace+"/"+o.Name]; ok {
				return dep
			}
			return o.Name
		case "StatefulSet", "DaemonSet", "Job", "CronJob":
			return o.Name
		}
	}
	return ""
}

// ── parsing helpers ──

func archLabel(arch string) string {
	arch = strings.ToLower(arch)
	switch {
	case strings.Contains(arch, "arm") || strings.Contains(arch, "aarch"):
		return "ARM64"
	case strings.Contains(arch, "amd") || strings.Contains(arch, "x86"):
		return "AMD64"
	default:
		return arch
	}
}

// parseCores parses a Kubernetes cpu quantity ("4", "1600m").
func parseCores(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if strings.HasSuffix(s, "m") {
		if v, err := strconv.ParseFloat(strings.TrimSuffix(s, "m"), 64); err == nil {
			return v / 1000
		}
		return 0
	}
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v
	}
	return 0
}

// parseBytes parses a Kubernetes memory quantity ("25116016640", "24576Ki").
func parseBytes(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	switch {
	case strings.HasSuffix(s, "Ki"):
		if v, err := strconv.ParseFloat(strings.TrimSuffix(s, "Ki"), 64); err == nil {
			return v * 1024
		}
	case strings.HasSuffix(s, "Mi"):
		if v, err := strconv.ParseFloat(strings.TrimSuffix(s, "Mi"), 64); err == nil {
			return v * 1024 * 1024
		}
	case strings.HasSuffix(s, "Gi"):
		if v, err := strconv.ParseFloat(strings.TrimSuffix(s, "Gi"), 64); err == nil {
			return v * 1024 * 1024 * 1024
		}
	default:
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			return v
		}
	}
	return 0
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	return time.Time{}
}

// providerFromID extracts the hosting provider from spec.providerID, falling
// back to a name-prefix guess when providerID is empty.
//   - "ocid1.instance..." -> "OCI Cloud"
//   - "metal://..."        -> "Edge"
//   - "proxmox://..."      -> "Proxmox"
func providerFromID(id, name string) string {
	switch {
	case strings.HasPrefix(id, "ocid1."):
		return "OCI Cloud"
	case strings.HasPrefix(id, "metal://"):
		return "Edge"
	case strings.Contains(id, "proxmox"):
		return "Proxmox"
	}
	switch {
	case strings.HasPrefix(name, "talos-oci") || strings.HasPrefix(name, "talosoci"):
		return "OCI Cloud"
	case strings.HasPrefix(name, "talosedge"):
		return "Edge"
	case strings.HasPrefix(name, "talos-os-") || strings.HasPrefix(name, "talos-amd64"):
		return "Cloud"
	default:
		return ""
	}
}

// tlsConfigFromCA builds a TLS config that trusts the in-cluster CA.
// Returns an error (instead of falling back to InsecureSkipVerify) if the CA
// cert cannot be read — fail closed.
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
