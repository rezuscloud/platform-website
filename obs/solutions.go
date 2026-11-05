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

// SolutionCategory is one capability class (a blueprint row) inside a tier,
// or a hosted-workload purpose.
type SolutionCategory struct {
	ID     string
	Label  string
	Blocks []WorkloadBlock
}

// Solution is a top-level presentation group of the live section: the
// blueprint's tiers (substrate / guarantees / planes) plus the hosted
// workloads the cloud serves. The Solution itself is unique — this split is
// presentation only, mirroring the cloud definition contract.
type Solution struct {
	ID         string
	Title      string
	Subtitle   string
	Categories []SolutionCategory
	Workloads  int // total blocks
	Instances  int // total per-host instances
}

// namespaceClass maps a namespace to its (tier, capability class) pair,
// following the blueprint rows of the cloud definition contract
// (wiki/concepts/cloud-contract.md): A compute, B network, C traffic entry,
// D block storage, E object storage, F image distribution, I DNS,
// G identity, H secrets, K backup, J observability, L consumption, M
// lifecycle. Namespaces absent from the map fall back to hosted workloads.
var namespaceClass = map[string][2]string{
	// ── Tier 1 · Substrate ──
	"kube-system":         {"substrate", "compute"}, // orchestration fabric (overrides below for CNI/DNS)
	"gateway-system":      {"substrate", "traffic"}, // Gateway API — row C
	"juicefs-csi":         {"substrate", "storage"},
	"juicefs-system":      {"substrate", "storage"},
	"juicefs-operator":    {"substrate", "storage"},
	"tikv-system":         {"substrate", "storage"}, // legacy juicefs metadata ns
	"local-path-storage":  {"substrate", "storage"},
	"snapshot-controller": {"substrate", "storage"}, // kubernetes-csi external-snapshotter
	"minio":               {"substrate", "object"},
	"object-storage":      {"substrate", "object"},
	"registry":            {"substrate", "images"},
	"distribution":        {"substrate", "images"},
	"external-dns":        {"substrate", "dns"},

	// ── Tier 2 · Guarantees ──
	"cilium-secrets":   {"guarantees", "secrets"},
	"cert-manager":     {"guarantees", "secrets"},
	"external-secrets": {"guarantees", "secrets"},
	"authentik":        {"guarantees", "identity"},
	"velero":           {"guarantees", "backup"},
	"signoz":           {"guarantees", "observability"},
	"monitoring":       {"guarantees", "observability"},
	"obi":              {"guarantees", "observability"},
	"telegraf":         {"guarantees", "observability"},

	// ── Tier 3 · Planes ──
	"rezuscloud":    {"planes", "consumption"}, // tenants, REST API, WebUI — row L
	"flux-system":   {"planes", "lifecycle"},   // GitOps delivery — row M
	"harmostes":     {"planes", "lifecycle"},
	"harmostes-dev": {"planes", "lifecycle"},
	"reloader":      {"planes", "lifecycle"}, // automated restart on config change

	// ── Hosted workloads (what the cloud serves — not contract rows) ──
	"forgejo":          {"workloads", "code"},
	"forgejo-runner":   {"workloads", "code"},
	"codeberg-runner":  {"workloads", "code"},
	"arc-runners":      {"workloads", "code"},
	"arc-systems":      {"workloads", "code"},
	"gitlab-runner":    {"workloads", "code"},
	"opencloud":        {"workloads", "files"},
	"collabora":        {"workloads", "files"},
	"searxng":          {"workloads", "search"},
	"vllm":             {"workloads", "ai"},
	"llamacpp":         {"workloads", "ai"},
	"dapr-system":      {"workloads", "runtime"},
	"keda":             {"workloads", "runtime"},
	"vela-system":      {"workloads", "runtime"},
	"cnpg-system":      {"workloads", "data"},
	"platform-website": {"workloads", "web"},
}

// nameClass overrides the namespace mapping for specific workload names.
var nameClass = map[string][2]string{
	"coredns-managed": {"substrate", "dns"}, // cluster DNS: row I
	"cilium-operator": {"substrate", "network"},
	"hubble-relay":    {"substrate", "network"},
	// metrics-server serves the resource-metrics API feeding scheduling and
	// HPA — part of the orchestration fabric (row A), not observability.
	"metrics-server": {"substrate", "compute"},
	"kube-prometheus-stack-kube-state-metrics": {"guarantees", "observability"},
	"local-path-provisioner":                   {"substrate", "storage"},
}

// solutionOrder fixes the rendering order of the presentation groups: the
// blueprint's three tiers, then the hosted workloads the cloud serves.
var solutionOrder = []string{"substrate", "guarantees", "planes", "workloads"}

// categoryOrder fixes the rendering order of capability classes within a
// tier (contract row order), and of workload purposes.
var categoryOrder = map[string][]string{
	"substrate":  {"compute", "network", "traffic", "storage", "object", "images", "dns"},
	"guarantees": {"identity", "secrets", "backup", "observability"},
	"planes":     {"consumption", "lifecycle"},
	"workloads":  {"code", "files", "search", "ai", "web", "runtime", "data", "apps"},
}

// categoryLabels name categories by their blueprint capability class
// (contract row letter), or by purpose for hosted workloads.
var categoryLabels = map[string]string{
	"compute":       "A · Compute & orchestration",
	"network":       "B · Network",
	"traffic":       "C · Traffic entry & load balancing",
	"storage":       "D · Block storage",
	"object":        "E · Object storage (S3)",
	"images":        "F · Image & artifact distribution",
	"dns":           "I · DNS",
	"identity":      "G · Identity & access",
	"secrets":       "H · Secrets & certificate lifecycle",
	"backup":        "K · Backup & DR",
	"observability": "J · Observability & measurement",
	"consumption":   "L · Consumption plane",
	"lifecycle":     "M · Lifecycle machinery",
	"code":          "Code & CI",
	"files":         "Files",
	"search":        "Search",
	"ai":            "AI",
	"web":           "Web",
	"runtime":       "App runtime",
	"data":          "Databases",
	"apps":          "Applications",
}

var solutionMeta = map[string][2]string{
	"substrate":  {"TIER 1 · SUBSTRATE", "abstraction & pooling — capability classes A–F, I"},
	"guarantees": {"TIER 2 · GUARANTEES", "identity, durability, observability — rows G–K, mandatory in every blueprint"},
	"planes":     {"TIER 3 · PLANES", "consumption & lifecycle — rows L–M, the operator removed"},
	"workloads":  {"HOSTED WORKLOADS", "what this cloud serves — inventory, not contract rows"},
}

// Classify returns the tier (or workloads group) and category for a
// workload. The name wins over the namespace when both are known.
func Classify(namespace, name string) (solution, category string) {
	if c, ok := nameClass[name]; ok {
		return c[0], c[1]
	}
	if c, ok := namespaceClass[namespace]; ok {
		return c[0], c[1]
	}
	return "workloads", "apps"
}

// BuildSolutions groups the live services into the contract's presentation
// tree: every workload (namespace/name) becomes exactly one block containing
// its per-host instances. Blocks land in one capability class of one tier —
// resources are unique, grouping is two-level.
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
