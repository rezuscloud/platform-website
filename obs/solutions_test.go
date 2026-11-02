package obs

import "testing"

func TestClassify(t *testing.T) {
	cases := []struct {
		ns, name, wantSol, wantCat string
	}{
		{"authentik", "authentik-server", "software", "identity"},
		{"kube-system", "cilium-operator", "infrastructure", "network"},
		{"kube-system", "coredns-managed", "infrastructure", "dns"},      // name wins over namespace
		{"kube-system", "metrics-server", "software", "observability"},  // platform service, not core
		{"forgejo", "forgejo", "software", "code"},
		{"unknown-namespace", "whatever", "software", "delivery"}, // default
		{"velero", "velero", "software", "backup"},
		{"flux-system", "helm-controller", "software", "delivery"},
		{"snapshot-controller", "snapshot-controller", "infrastructure", "storage"}, // kubernetes-csi
		{"cnpg-system", "cloudnative-pg", "software", "data"},
		{"platform-website", "platform-website", "software", "web"},
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

	if len(sols) != 2 {
		t.Fatalf("got %d solutions, want 2", len(sols))
	}
	infra, software := sols[0], sols[1]
	if infra.ID != "infrastructure" || software.ID != "software" {
		t.Fatalf("order = %s,%s want infrastructure,software", infra.ID, software.ID)
	}

	// infrastructure: network (1 workload/1 instance) then storage (1/3)
	if len(infra.Categories) < 2 || infra.Categories[0].ID != "network" || infra.Categories[1].ID != "storage" {
		t.Fatalf("infra categories = %+v", infra.Categories)
	}
	if len(infra.Categories[1].Blocks) != 1 || len(infra.Categories[1].Blocks[0].Instances) != 3 {
		t.Errorf("storage block should contain 3 instances, got %+v", infra.Categories[1].Blocks)
	}
	if infra.Workloads != 2 || infra.Instances != 4 {
		t.Errorf("infra counts = %d/%d, want 2/4", infra.Workloads, infra.Instances)
	}

	// software: identity and code categories; forgejo appears once with 2 hosts
	var codeCat, idCat *SolutionCategory
	for i := range software.Categories {
		switch software.Categories[i].ID {
		case "code":
			codeCat = &software.Categories[i]
		case "identity":
			idCat = &software.Categories[i]
		}
	}
	if codeCat == nil || idCat == nil {
		t.Fatalf("software categories missing: %+v", software.Categories)
	}
	if len(codeCat.Blocks) != 1 || codeCat.Blocks[0].Name != "forgejo" || len(codeCat.Blocks[0].Instances) != 2 {
		t.Errorf("forgejo block = %+v", codeCat.Blocks)
	}
	if software.Workloads != 2 || software.Instances != 3 {
		t.Errorf("software counts = %d/%d, want 2/3", software.Workloads, software.Instances)
	}
}

func TestBuildSolutionsEveryResourceUnique(t *testing.T) {
	data := LiveData{Services: []Service{
		{Namespace: "a", Name: "x", Host: "h1"},
		{Namespace: "a", Name: "x", Host: "h2"},
		{Namespace: "a", Name: "x", Host: "h3"},
	}}
	sols := BuildSolutions(data)
	total := 0
	for _, s := range sols {
		for _, c := range s.Categories {
			total += len(c.Blocks)
		}
	}
	if total != 1 {
		t.Errorf("same workload on 3 hosts must yield 1 block, got %d", total)
	}
}
