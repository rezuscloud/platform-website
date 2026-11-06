package obs

import "fmt"

// WorkloadBlock groups every per-host instance of one workload (namespace/name)
// into a single block. Instances are the per-host Service records carried by
// the existing SSE stream — the block is a view over them, not new data.
type WorkloadBlock struct {
	Namespace string
	Name      string
	Instances []Service
}

// Solution is one presentation group of the live section: one of the six
// contracts that form a cloud (cloud definition contract), holding its
// workloads flat — no capability-class middle layer. The Solution itself is
// unique; this split is presentation only.
type Solution struct {
	ID        string
	Title     string
	Subtitle  string
	Blocks    []WorkloadBlock
	Workloads int // total blocks
	Instances int // total per-host instances
}

// contractOrder fixes the rendering order: the six contracts, numbered as in
// the cloud definition contract's table.
var contractOrder = []string{"consumption", "lifecycle", "pooling", "identity", "durability", "observability"}

// contractMeta carries the name and the guarantee each contract delivers —
// one line, straight from the contract.
var contractMeta = map[string][2]string{
	"consumption":   {"CONSUMPTION", "capacity and services obtained without the operator — self-served"},
	"lifecycle":     {"LIFECYCLE", "the cloud outlives its parts: nodes die, versions bump, certs expire — it reconciles itself"},
	"pooling":       {"ABSTRACTION & POOLING", "capability offered as classes, not machines"},
	"identity":      {"IDENTITY & TENANCY", "who may do what — identity plane, authz, consumer boundaries"},
	"durability":    {"DURABILITY", "data survives failure, workloads survive node loss"},
	"observability": {"OBSERVABILITY & MEASUREMENT", "operator and consumer can see the cloud"},
}

// namespaceContract maps a namespace to the contract its workload delivers.
// Names absent from nameContract and namespaces absent from this map fall
// back to consumption: unknown workloads are things the cloud serves.
var namespaceContract = map[string]string{
	// 1 · Consumption — the plane (tenants, REST API, WebUI) and everything
	// it serves: hosted workloads are what consumers obtain.
	"rezuscloud":       "consumption",
	"forgejo":          "consumption",
	"forgejo-runner":   "consumption",
	"codeberg-runner":  "consumption",
	"arc-runners":      "consumption",
	"arc-systems":      "consumption",
	"gitlab-runner":    "consumption",
	"opencloud":        "consumption",
	"collabora":        "consumption",
	"searxng":          "consumption",
	"vllm":             "consumption",
	"llamacpp":         "consumption",
	"dapr-system":      "consumption",
	"keda":             "consumption",
	"vela-system":      "consumption",
	"cnpg-system":      "consumption",
	"platform-website": "consumption",

	// 2 · Lifecycle — declarative reconciliation and delivery machinery.
	"flux-system":   "lifecycle",
	"harmostes":     "lifecycle",
	"harmostes-dev": "lifecycle",
	"reloader":      "lifecycle",

	// 3 · Abstraction & pooling — the standardized core, offered as classes.
	"kube-system":        "pooling",
	"gateway-system":     "pooling",
	"juicefs-csi":        "pooling",
	"juicefs-system":     "pooling",
	"juicefs-operator":   "pooling",
	"tikv-system":        "pooling", // legacy juicefs metadata ns
	"local-path-storage": "pooling",
	"minio":              "pooling",
	"object-storage":     "pooling",
	"registry":           "pooling",
	"distribution":       "pooling",
	"external-dns":       "pooling",

	// 4 · Identity & tenancy — identity plane, secrets and certificate trust.
	"authentik":        "identity",
	"cert-manager":     "identity",
	"external-secrets": "identity",
	"cilium-secrets":   "identity",

	// 5 · Durability — what makes data and workloads survive failure.
	"velero":              "durability",
	"snapshot-controller": "durability", // CSI volume snapshots

	// 6 · Observability & measurement.
	"signoz":     "observability",
	"monitoring": "observability",
	"obi":        "observability",
	"telegraf":   "observability",
}

// nameContract overrides the namespace mapping for specific workload names.
var nameContract = map[string]string{
	"coredns-managed": "pooling", // cluster DNS
	"cilium-operator": "pooling",
	"hubble-relay":    "pooling",
	// metrics-server serves the resource-metrics API feeding scheduling and
	// HPA — part of the orchestration fabric.
	"metrics-server":                           "pooling",
	"local-path-provisioner":                   "pooling",
	"kube-prometheus-stack-kube-state-metrics": "observability",
}

// Classify returns the contract a workload delivers. The name wins over the
// namespace when both are known; unknown workloads fall back to consumption.
func Classify(namespace, name string) string {
	if c, ok := nameContract[name]; ok {
		return c
	}
	if c, ok := namespaceContract[namespace]; ok {
		return c
	}
	return "consumption"
}

// BuildSolutions groups the live services by the six contracts that form the
// cloud: every workload (namespace/name) becomes exactly one block containing
// its per-host instances — resources are unique, grouping is one level.
func BuildSolutions(data LiveData) []Solution {
	type key struct{ ns, name string }
	blocks := map[key]*WorkloadBlock{}
	order := []key{}
	for _, svc := range data.Services {
		k := key{svc.Namespace, svc.Name}
		if _, ok := blocks[k]; !ok {
			blocks[k] = &WorkloadBlock{Namespace: svc.Namespace, Name: svc.Name}
			order = append(order, k)
		}
		blocks[k].Instances = append(blocks[k].Instances, svc)
	}

	perContract := map[string][]WorkloadBlock{}
	counts := map[string][2]int{} // contract -> {workloads, instances}
	for _, k := range order {
		contract := Classify(k.ns, k.name)
		perContract[contract] = append(perContract[contract], *blocks[k])
		c := counts[contract]
		counts[contract] = [2]int{c[0] + 1, c[1] + len(blocks[k].Instances)}
	}

	solutions := make([]Solution, 0, len(contractOrder))
	for _, contract := range contractOrder {
		b, ok := perContract[contract]
		if !ok || len(b) == 0 {
			continue
		}
		meta := contractMeta[contract]
		solutions = append(solutions, Solution{
			ID: contract, Title: meta[0], Subtitle: meta[1],
			Blocks:    b,
			Workloads: counts[contract][0], Instances: counts[contract][1],
		})
	}
	return solutions
}

// CountLabel renders the workload/instance counts of a solution bar.
func CountLabel(s Solution) string {
	return fmt.Sprintf("%d workloads · %d instances", s.Workloads, s.Instances)
}
