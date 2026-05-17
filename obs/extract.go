package obs

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// BuildServices discovers services from CPU metric series and enriches
// them with RAM, disk, network, and sparkline data from all results.
func BuildServices(cpuSeries []v3Series, allResults map[string][]v3Series, now time.Time) []Service {
	type svcInfo struct {
		ns, deploy, host, uptime string
	}
	svcMap := map[string]*svcInfo{}

	for _, s := range cpuSeries {
		ns := LabelStr(s.Labels, "k8s_namespace_name", "k8s.namespace.name")
		deploy := LabelStr(s.Labels, "k8s.deployment.name")
		if deploy == "" {
			deploy = LabelStr(s.Labels, "k8s.statefulset.name")
		}
		if deploy == "" {
			deploy = LabelStr(s.Labels, "k8s.daemonset.name")
		}
		if ns == "" || deploy == "" {
			continue
		}
		key := ns + "/" + deploy
		if _, exists := svcMap[key]; exists {
			continue
		}
		info := &svcInfo{ns: ns, deploy: deploy}
		svcMap[key] = info
		if t := LabelStr(s.Labels, "k8s.pod.start_time"); t != "" {
			info.uptime = t
		}
		if node := LabelStr(s.Labels, "k8s_node_name", "k8s.node.name"); node != "" {
			info.host = node
		}
	}

	deployCPU := LatestByDeployment(allResults["cpu"])
	deployRAM := LatestByDeployment(allResults["ram"])
	deployDisk := LatestByDeployment(allResults["disk"])
	deployNet := LatestByDeployment(allResults["net"])

	sparkCPU := SparkByDeployment(allResults["cpu"])
	sparkRAM := SparkByDeployment(allResults["ram"])
	sparkDisk := SparkByDeployment(allResults["disk"])
	sparkNet := SparkByDeployment(allResults["net"])

	services := make([]Service, 0, len(svcMap))
	for _, info := range svcMap {
		key := info.ns + "/" + info.deploy
		svc := Service{
			Name:      info.deploy,
			Namespace: info.ns,
			Category:  CategoryForNamespace(info.ns),
			Host:      info.host,
			Status:    "running",
		}
		if info.uptime != "" {
			if parsed, err := time.Parse(time.RFC3339, info.uptime); err == nil {
				svc.Uptime = FormatUptime(now.Sub(parsed))
			}
		}

		if v, ok := deployCPU[key]; ok && v <= 100 {
			svc.CPU = math.Round(v*1000) / 1000
		}
		if v, ok := deployRAM[key]; ok {
			svc.RAM = math.Round(v/1024/1024*10) / 10
		}
		if v, ok := deployNet[key]; ok {
			svc.NetKB = math.Round(v/1024*10) / 10
		}
		if v, ok := deployDisk[key]; ok {
			svc.DiskMB = math.Round(v/1024/1024*10) / 10
		}

		if pts, ok := sparkCPU[key]; ok {
			svc.CPUHist = pts
		}
		if pts, ok := sparkRAM[key]; ok {
			svc.RAMHist = pts
		}
		if pts, ok := sparkNet[key]; ok {
			svc.NetHist = pts
		}
		if pts, ok := sparkDisk[key]; ok {
			svc.DiskHist = pts
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

// BuildHosts extracts host data from node-level metric results.
func BuildHosts(results map[string][]v3Series) []Host {
	nodeCPU := LatestByNode(results["nodeCpu"])
	nodeRAM := LatestByNode(results["nodeRam"])
	nodeUp := LatestByNode(results["nodeUp"])

	nodeCount := map[string]int{}
	for _, s := range results["cpu"] {
		if node := LabelStr(s.Labels, "k8s_node_name", "k8s.node.name"); node != "" {
			nodeCount[node]++
		}
	}

	nodeNames := DiscoverNodeNames(nodeCPU, nodeRAM, nodeUp, nodeCount)

	hosts := make([]Host, 0, len(nodeNames))
	for _, name := range nodeNames {
		h := Host{Name: name}
		if strings.Contains(name, "control-plane") {
			h.Label = "Cloud"
			h.Detail = "Control plane"
		} else {
			h.Label = "Edge"
			h.Detail = "Worker node"
		}
		if v, ok := nodeCPU[name]; ok {
			h.CPU = math.Round(v*100) / 100
		}
		if v, ok := nodeRAM[name]; ok {
			h.RAM = math.Round(v/1024/1024*10) / 10
		}
		if v, ok := nodeUp[name]; ok {
			h.Uptime = FormatUptime(time.Duration(v) * time.Second)
		}
		if n, ok := nodeCount[name]; ok {
			h.SvcCount = n
		}
		hosts = append(hosts, h)
	}
	return hosts
}

// LatestByDeployment returns map[ns/deploy]latestValue using k8s.deployment.name label.
// Takes max across pods of the same deployment.
func LatestByDeployment(series []v3Series) map[string]float64 {
	m := map[string]float64{}
	for _, s := range series {
		ns := LabelStr(s.Labels, "k8s_namespace_name", "k8s.namespace.name")
		deploy := LabelStr(s.Labels, "k8s.deployment.name")
		if deploy == "" {
			deploy = LabelStr(s.Labels, "k8s.statefulset.name")
		}
		if deploy == "" {
			deploy = LabelStr(s.Labels, "k8s.daemonset.name")
		}
		if ns == "" || deploy == "" || len(s.Values) == 0 {
			continue
		}
		key := ns + "/" + deploy
		last := s.Values[len(s.Values)-1].Value
		v, err := parseFloat(last)
		if err != nil {
			continue
		}
		if existing, ok := m[key]; ok {
			if v > existing {
				m[key] = v
			}
		} else {
			m[key] = v
		}
	}
	return m
}

// SparkByDeployment returns map[ns/deploy]svgPolyline using k8s.deployment.name label.
func SparkByDeployment(series []v3Series) map[string]string {
	m := map[string]string{}
	for _, s := range series {
		ns := LabelStr(s.Labels, "k8s_namespace_name", "k8s.namespace.name")
		deploy := LabelStr(s.Labels, "k8s.deployment.name")
		if deploy == "" {
			deploy = LabelStr(s.Labels, "k8s.statefulset.name")
		}
		if deploy == "" {
			deploy = LabelStr(s.Labels, "k8s.daemonset.name")
		}
		if ns == "" || deploy == "" || len(s.Values) < 1 {
			continue
		}
		key := ns + "/" + deploy
		if _, exists := m[key]; exists {
			continue
		}
		vals := make([]float64, len(s.Values))
		for i, p := range s.Values {
			vals[i], _ = parseFloat(p.Value)
		}
		for i, v := range vals {
			if v > 100 {
				vals[i] = 0
			}
		}
		if len(vals) == 1 {
			vals = []float64{vals[0], vals[0]}
		}
		m[key] = sparklinePoints(vals, 48, 16)
	}
	return m
}

func LatestByNode(series []v3Series) map[string]float64 {
	m := map[string]float64{}
	for _, s := range series {
		node := s.Labels["k8s.node.name"]
		if node == "" || len(s.Values) == 0 {
			continue
		}
		last := s.Values[len(s.Values)-1].Value
		if v, err := parseFloat(last); err == nil {
			m[node] = v
		}
	}
	return m
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
