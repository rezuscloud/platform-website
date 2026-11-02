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

// SolutionCategory is one capability group inside a solution.
type SolutionCategory struct {
	ID     string
	Label  string
	Blocks []WorkloadBlock
}

// Solution is a top-level group of the live section. The minimum model is two:
// the infrastructure (the cloud itself) and the software running on top.
type Solution struct {
	ID         string
	Title      string
	Subtitle   string
	Categories []SolutionCategory
	Workloads  int // total blocks
	Instances  int // total per-host instances
}

// namespaceClass maps a namespace to its (solution, category) pair.
// Namespaces absent from the map fall back to infrastructure/delivery.
var namespaceClass = map[string][2]string{
	// infrastructure — standardized cluster core (CNI, CSI, DNS, secrets)
	"kube-system":         {"infrastructure", "network"},
	"gateway-system":      {"infrastructure", "network"},
	"juicefs-csi":         {"infrastructure", "storage"},
	"juicefs-system":      {"infrastructure", "storage"},
	"juicefs-operator":    {"infrastructure", "storage"},
	"tikv-system":         {"infrastructure", "storage"}, // legacy juicefs metadata ns
	"local-path-storage":  {"infrastructure", "storage"},
	"snapshot-controller": {"infrastructure", "storage"}, // kubernetes-csi external-snapshotter
	"cilium-secrets":      {"infrastructure", "secrets"},
	"cert-manager":        {"infrastructure", "secrets"},
	"external-secrets":    {"infrastructure", "secrets"},

	// software — everything running on top of the cluster
	"external-dns":     {"software", "dns"},
	"signoz":           {"software", "observability"},
	"monitoring":       {"software", "observability"},
	"obi":              {"software", "observability"},
	"telegraf":         {"software", "observability"},
	"velero":           {"software", "backup"},
	"flux-system":      {"software", "delivery"},
	"vela-system":      {"software", "delivery"},
	"dapr-system":      {"software", "delivery"},
	"keda":             {"software", "delivery"},
	"reloader":         {"software", "delivery"},
	"cnpg-system":      {"software", "data"},
	"authentik":        {"software", "identity"},
	"forgejo":          {"software", "code"},
	"forgejo-runner":   {"software", "code"},
	"codeberg-runner":  {"software", "code"},
	"arc-runners":      {"software", "code"},
	"arc-systems":      {"software", "code"},
	"gitlab-runner":    {"software", "code"},
	"opencloud":        {"software", "files"},
	"collabora":        {"software", "files"},
	"searxng":          {"software", "search"},
	"vllm":             {"software", "ai"},
	"llamacpp":         {"software", "ai"},
	"harmostes":        {"software", "automation"},
	"harmostes-dev":    {"software", "automation"},
	"platform-website": {"software", "web"},
	"rezuscloud":       {"software", "web"},
}

// nameClass overrides the namespace mapping for specific workload names.
var nameClass = map[string][2]string{
	"coredns-managed": {"infrastructure", "dns"}, // cluster DNS: standardized core
	"cilium-operator": {"infrastructure", "network"},
	"hubble-relay":    {"infrastructure", "network"},
	// metrics stack serves the HPA/metrics APIs but is a platform service
	"metrics-server": {"software", "observability"},
	"kube-prometheus-stack-kube-state-metrics": {"software", "observability"},
	"local-path-provisioner":                   {"infrastructure", "storage"},
}

// solutionOrder fixes the rendering order of solutions.
var solutionOrder = []string{"infrastructure", "software"}

// categoryOrder fixes the rendering order of categories within a solution.
var categoryOrder = map[string][]string{
	"infrastructure": {"network", "storage", "dns", "secrets"},
	"software":       {"identity", "code", "files", "search", "ai", "automation", "web", "data", "observability", "backup", "delivery"},
}

var categoryLabels = map[string]string{
	"network": "Network · CNI", "storage": "Storage · CSI", "dns": "DNS", "secrets": "Secrets",
	"observability": "Observability", "backup": "Backup", "delivery": "Delivery", "data": "Data",
	"identity": "Identity", "code": "Code", "files": "Files", "search": "Search",
	"ai": "AI", "automation": "Automation", "web": "Web",
}

var solutionMeta = map[string][2]string{
	"infrastructure": {"INFRASTRUCTURE", "standardized cluster core"},
	"software":       {"SOFTWARE", "everything running on top"},
}

// Classify returns the solution and category for a workload. The name wins
// over the namespace when both are known.
func Classify(namespace, name string) (solution, category string) {
	if c, ok := nameClass[name]; ok {
		return c[0], c[1]
	}
	if c, ok := namespaceClass[namespace]; ok {
		return c[0], c[1]
	}
	return "software", "delivery"
}

// BuildSolutions groups the live services into the two-solution tree: every
// workload (namespace/name) becomes exactly one block containing its per-host
// instances. Blocks land in one category of one solution — resources are
// unique, grouping is two-level.
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

	perSol := map[string]map[string][]WorkloadBlock{}
	counts := map[string][2]int{} // solution -> {workloads, instances}
	for _, k := range order {
		solID, catID := Classify(k.ns, k.name)
		if perSol[solID] == nil {
			perSol[solID] = map[string][]WorkloadBlock{}
		}
		perSol[solID][catID] = append(perSol[solID][catID], *blocks[k])
		c := counts[solID]
		counts[solID] = [2]int{c[0] + 1, c[1] + len(blocks[k].Instances)}
	}

	solutions := make([]Solution, 0, len(solutionOrder))
	for _, solID := range solutionOrder {
		cats, ok := perSol[solID]
		if !ok {
			continue
		}
		meta := solutionMeta[solID]
		s := Solution{
			ID: solID, Title: meta[0], Subtitle: meta[1],
			Workloads: counts[solID][0], Instances: counts[solID][1],
		}
		for _, catID := range categoryOrder[solID] {
			blocksInCat, ok := cats[catID]
			if !ok || len(blocksInCat) == 0 {
				continue
			}
			s.Categories = append(s.Categories, SolutionCategory{
				ID: catID, Label: categoryLabels[catID], Blocks: blocksInCat,
			})
		}
		solutions = append(solutions, s)
	}
	return solutions
}

// CategoryCountLabel renders the workload count of a category.
func CategoryCountLabel(c SolutionCategory) string {
	if len(c.Blocks) == 1 {
		return "1 workload"
	}
	return fmt.Sprintf("%d workloads", len(c.Blocks))
}
