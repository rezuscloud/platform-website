package obs

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// FormatUptime returns a human-readable duration.
func FormatUptime(d time.Duration) string {
	h := int(d.Hours())
	if h >= 24 {
		return fmt.Sprintf("%dd", h/24)
	}
	if h >= 1 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}

// WorkloadKey extracts a unique "namespace/workload-name" key from SigNoz
// metric labels. Tries deployment, statefulset, and daemonset labels in
// order. Returns empty string if no workload or namespace is found.
// Single place to fix if SigNoz label keys change.
func WorkloadKey(labels map[string]string) string {
	ns := LabelStr(labels, "k8s_namespace_name", "k8s.namespace.name")
	deploy := LabelStr(labels, "k8s.deployment.name")
	if deploy == "" {
		deploy = LabelStr(labels, "k8s.statefulset.name")
	}
	if deploy == "" {
		deploy = LabelStr(labels, "k8s.daemonset.name")
	}
	if ns == "" || deploy == "" {
		return ""
	}
	return ns + "/" + deploy
}

// BuildServices converts a MetricsSnapshot into a sorted, filtered Service list.
func BuildServices(snap MetricsSnapshot, now time.Time) []Service {
	services := make([]Service, 0, len(snap.Workloads))
	for key, wm := range snap.Workloads {
		_ = key
		svc := Service{
			Name:      wm.Name,
			Namespace: wm.Namespace,
			Category:  CategoryForNamespace(wm.Namespace),
			Host:      wm.Host,
			Status:    "running",
		}
		if wm.Uptime != "" {
			if parsed, err := time.Parse(time.RFC3339, wm.Uptime); err == nil {
				svc.Uptime = FormatUptime(now.Sub(parsed))
			}
		}

		if wm.CPU > 0 {
			svc.CPU = wm.CPU
		}
		if wm.RAM > 0 {
			svc.RAM = math.Round(wm.RAM/1024/1024*10) / 10
		}
		if wm.Net > 0 {
			svc.NetKB = math.Round(wm.Net/1024*10) / 10
		}
		if wm.Disk > 0 {
			svc.DiskMB = math.Round(wm.Disk/1024/1024*10) / 10
		}

		if len(wm.CPUHist) > 0 {
			svc.CPUHist = SparklinePoints(wm.CPUHist, 48, 16)
		}
		if len(wm.RAMHist) > 0 {
			svc.RAMHist = SparklinePoints(wm.RAMHist, 48, 16)
		}
		if len(wm.NetHist) > 0 {
			svc.NetHist = SparklinePoints(wm.NetHist, 48, 16)
		}
		if len(wm.DiskHist) > 0 {
			svc.DiskHist = SparklinePoints(wm.DiskHist, 48, 16)
		}

		services = append(services, svc)
	}

	// Filter stale services: pods from deleted namespaces with no active metrics.
	active := make([]Service, 0, len(services))
	for _, svc := range services {
		if svc.CPU > 0 || svc.RAM > 0 || svc.NetKB > 0 || svc.DiskMB > 0 {
			active = append(active, svc)
		}
	}

	SortServices(active)
	return active
}

// BuildHosts converts a MetricsSnapshot into a sorted Host list.
func BuildHosts(snap MetricsSnapshot) []Host {
	nodeNames := make([]string, 0, len(snap.Nodes))
	for name := range snap.Nodes {
		nodeNames = append(nodeNames, name)
	}
	sort.Slice(nodeNames, func(i, j int) bool {
		ci := strings.Contains(nodeNames[i], "control-plane")
		cj := strings.Contains(nodeNames[j], "control-plane")
		if ci != cj {
			return ci
		}
		return nodeNames[i] < nodeNames[j]
	})

	hosts := make([]Host, 0, len(nodeNames))
	for _, name := range nodeNames {
		nm := snap.Nodes[name]
		h := Host{Name: name}
		if strings.Contains(name, "control-plane") {
			h.Label = "Cloud"
			h.Detail = "Control plane"
		} else {
			h.Label = "Edge"
			h.Detail = "Worker node"
		}
		if nm.CPU > 0 {
			h.CPU = math.Round(nm.CPU*100) / 100
		}
		if nm.RAM > 0 {
			h.RAM = math.Round(nm.RAM/1024/1024*10) / 10
		}
		if nm.Uptime > 0 {
			h.Uptime = FormatUptime(time.Duration(nm.Uptime) * time.Second)
		}
		if n, ok := snap.NodeSvcCounts[name]; ok {
			h.SvcCount = n
		}
		hosts = append(hosts, h)
	}
	return hosts
}

// SparklinePoints returns just the polyline points (no area) from Sparkline.
func SparklinePoints(values []float64, w, h int) string {
	pts, _ := Sparkline(values, w, h)
	return pts
}
