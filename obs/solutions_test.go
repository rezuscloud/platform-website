package obs

import "testing"

func TestClassify(t *testing.T) {
	cases := []struct {
		ns, name, wantSol, wantCat string
	}{
		// Tier 1 · Substrate
		{"kube-system", "cilium-operator", "substrate", "network"}, // name wins over namespace
		{"kube-system", "coredns-managed", "substrate", "dns"},     // name wins over namespace
		{"kube-system", "metrics-server", "substrate", "compute"},  // resource-metrics API feeds scheduling/HPA
		{"kube-system", "kube-apiserver", "substrate", "compute"},
		{"gateway-system", "gateway-api", "substrate", "traffic"}, // row C
		{"juicefs-csi", "juicefs-csi-node", "substrate", "storage"},
		{"snapshot-controller", "snapshot-controller", "substrate", "storage"}, // kubernetes-csi
		{"external-dns", "external-dns", "substrate", "dns"},
		// Tier 2 · Guarantees
		{"authentik", "authentik-server", "guarantees", "identity"},        // row G
		{"cert-manager", "cert-manager", "guarantees", "secrets"},          // row H
		{"external-secrets", "external-secrets", "guarantees", "secrets"},  // row H
		{"velero", "velero", "guarantees", "backup"},                       // row K
		{"signoz", "signoz-otel-collector", "guarantees", "observability"}, // row J
		// Tier 3 · Planes
		{"rezuscloud", "rezuscloud", "planes", "consumption"},     // row L
		{"flux-system", "helm-controller", "planes", "lifecycle"}, // row M
		{"harmostes", "harmostes", "planes", "lifecycle"},         // row M
		// Hosted workloads
		{"forgejo", "forgejo", "workloads", "code"},
		{"cnpg-system", "cloudnative-pg", "workloads", "data"},
		{"platform-website", "platform-website", "workloads", "web"},
		{"dapr-system", "dapr-sidecar", "workloads", "runtime"},
		{"unknown-namespace", "whatever", "workloads", "apps"}, // default
	}
	for _, c := range cases {
		sol, cat := Classify(c.ns, c.name)
		if sol != c.wantSol || cat != c.wantCat {
			t.Errorf("Classify(%q,%q) = %s/%s, want %s/%s", c.ns, c.name, sol, cat, c.wantSol, c.wantCat)
		}
	}
}

func TestBuildSolutionsGroupsPerHostInstances(t *testing.T) {
	data := LiveData{Services: []Service{
		{Namespace: "forgejo", Name: "forgejo", Host: "h1", Status: "running"},
		{Namespace: "forgejo", Name: "forgejo", Host: "h2", Status: "running"},
		{Namespace: "authentik", Name: "authentik-server", Host: "h1", Status: "running"},
		{Namespace: "kube-system", Name: "cilium-operator", Host: "h1", Status: "running"},
		{Namespace: "juicefs-csi", Name: "juicefs-csi-node", Host: "h1", Status: "running"},
		{Namespace: "juicefs-csi", Name: "juicefs-csi-node", Host: "h2", Status: "running"},
		{Namespace: "juicefs-csi", Name: "juicefs-csi-node", Host: "h3", Status: "running"},
	}}
	sols := BuildSolutions(data)

	// planes has no services → skipped; order is substrate, guarantees, workloads
	if len(sols) != 3 {
		t.Fatalf("got %d solutions, want 3", len(sols))
	}
	substrate, guarantees, workloads := sols[0], sols[1], sols[2]
	if substrate.ID != "substrate" || guarantees.ID != "guarantees" || workloads.ID != "workloads" {
		t.Fatalf("order = %s,%s,%s want substrate,guarantees,workloads", substrate.ID, guarantees.ID, workloads.ID)
	}

	// substrate: network (row B, 1 workload/1 instance) then storage (row D, 1/3)
	if len(substrate.Categories) < 2 || substrate.Categories[0].ID != "network" || substrate.Categories[1].ID != "storage" {
		t.Fatalf("substrate categories = %+v", substrate.Categories)
	}
	if len(substrate.Categories[1].Blocks) != 1 || len(substrate.Categories[1].Blocks[0].Instances) != 3 {
		t.Errorf("storage block should contain 3 instances, got %+v", substrate.Categories[1].Blocks)
	}
	if substrate.Workloads != 2 || substrate.Instances != 4 {
		t.Errorf("substrate counts = %d/%d, want 2/4", substrate.Workloads, substrate.Instances)
	}

	// workloads: code and (empty categories skipped); forgejo appears once with 2 hosts
	var codeCat *SolutionCategory
	for i := range workloads.Categories {
		if workloads.Categories[i].ID == "code" {
			codeCat = &workloads.Categories[i]
		}
	}
	if codeCat == nil {
		t.Fatalf("workloads categories missing code: %+v", workloads.Categories)
	}
	if len(codeCat.Blocks) != 1 || len(codeCat.Blocks[0].Instances) != 2 {
		t.Errorf("code block should contain 1 workload with 2 instances, got %+v", codeCat.Blocks)
	}
	if workloads.Workloads != 1 || workloads.Instances != 2 {
		t.Errorf("workloads counts = %d/%d, want 1/2", workloads.Workloads, workloads.Instances)
	}

	// guarantees tier: identity (row G) carries authentik with its single instance
	if len(guarantees.Categories) != 1 || guarantees.Categories[0].ID != "identity" {
		t.Fatalf("guarantees categories = %+v", guarantees.Categories)
	}
	if guarantees.Workloads != 1 || guarantees.Instances != 1 {
		t.Errorf("guarantees counts = %d/%d, want 1/1", guarantees.Workloads, guarantees.Instances)
	}
}

func TestBuildSolutionsEveryResourceUnique(t *testing.T) {
	data := LiveData{Services: []Service{
		{Namespace: "authentik", Name: "authentik-server", Host: "h1", Status: "running"},
		{Namespace: "authentik", Name: "authentik-server", Host: "h2", Status: "running"},
	}}
	sols := BuildSolutions(data)

	var total int
	for _, s := range sols {
		for _, c := range s.Categories {
			total += len(c.Blocks)
			for _, b := range c.Blocks {
				if b.Namespace != "authentik" || b.Name != "authentik-server" {
					t.Errorf("unexpected block %+v", b)
				}
			}
		}
	}
	if total != 1 {
		t.Fatalf("same workload split across %d blocks, want exactly 1", total)
	}
}
