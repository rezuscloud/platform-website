package obs

import "testing"

func TestClassify(t *testing.T) {
	cases := []struct {
		ns, name, want string
	}{
		// 1 · Consumption — the plane and everything it serves
		{"rezuscloud", "rezuscloud", "consumption"},
		{"forgejo", "forgejo", "consumption"},
		{"forgejo", "forgejo-runner", "consumption"},
		{"opencloud", "opencloud", "consumption"},
		{"vllm", "vllm", "consumption"},
		{"platform-website", "platform-website", "consumption"},
		{"cnpg-system", "cloudnative-pg", "consumption"},
		{"dapr-system", "dapr-sidecar", "consumption"},
		{"unknown-namespace", "whatever", "consumption"}, // default: the cloud serves it
		// 2 · Lifecycle
		{"flux-system", "helm-controller", "lifecycle"},
		{"harmostes", "harmostes", "lifecycle"},
		{"reloader", "reloader", "lifecycle"},
		// 3 · Abstraction & pooling
		{"kube-system", "cilium-operator", "pooling"}, // name wins over namespace
		{"kube-system", "coredns-managed", "pooling"}, // name wins over namespace
		{"kube-system", "metrics-server", "pooling"},  // resource-metrics feeds scheduling/HPA
		{"gateway-system", "gateway-api", "pooling"},
		{"juicefs-csi", "juicefs-csi-node", "pooling"},
		{"local-path-storage", "local-path-provisioner", "pooling"},
		{"external-dns", "external-dns", "pooling"},
		// 4 · Identity & tenancy
		{"authentik", "authentik-server", "identity"},
		{"cert-manager", "cert-manager", "identity"},
		{"external-secrets", "external-secrets", "identity"},
		// 5 · Durability
		{"velero", "velero", "durability"},
		{"snapshot-controller", "snapshot-controller", "durability"}, // CSI volume snapshots
		// 6 · Observability & measurement
		{"signoz", "signoz-otel-collector", "observability"},
		{"monitoring", "prometheus", "observability"},
		{"monitoring", "kube-prometheus-stack-kube-state-metrics", "observability"}, // name wins
	}
	for _, c := range cases {
		if got := Classify(c.ns, c.name); got != c.want {
			t.Errorf("Classify(%q,%q) = %s, want %s", c.ns, c.name, got, c.want)
		}
	}
}

func TestBuildSolutionsGroupsFlatByContract(t *testing.T) {
	data := LiveData{Services: []Service{
		{Namespace: "forgejo", Name: "forgejo", Host: "h1", Status: "running"},
		{Namespace: "forgejo", Name: "forgejo", Host: "h2", Status: "running"},
		{Namespace: "authentik", Name: "authentik-server", Host: "h1", Status: "running"},
		{Namespace: "kube-system", Name: "cilium-operator", Host: "h1", Status: "running"},
		{Namespace: "juicefs-csi", Name: "juicefs-csi-node", Host: "h1", Status: "running"},
		{Namespace: "juicefs-csi", Name: "juicefs-csi-node", Host: "h2", Status: "running"},
		{Namespace: "juicefs-csi", Name: "juicefs-csi-node", Host: "h3", Status: "running"},
		{Namespace: "velero", Name: "velero", Host: "h1", Status: "running"},
	}}
	sols := BuildSolutions(data)

	// contracts with content only, in contract-table order
	var got []string
	for _, s := range sols {
		got = append(got, s.ID)
	}
	want := []string{"consumption", "pooling", "identity", "durability"}
	if len(got) != len(want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}

	byID := map[string]Solution{}
	for _, s := range sols {
		byID[s.ID] = s
	}

	// consumption: forgejo appears once with 2 hosts, flat (no categories)
	c := byID["consumption"]
	if len(c.Blocks) != 1 || len(c.Blocks[0].Instances) != 2 {
		t.Errorf("consumption blocks = %+v, want 1 block with 2 instances", c.Blocks)
	}
	if c.Workloads != 1 || c.Instances != 2 {
		t.Errorf("consumption counts = %d/%d, want 1/2", c.Workloads, c.Instances)
	}

	// pooling: cilium (1/1) then juicefs (1/3)
	p := byID["pooling"]
	if p.Workloads != 2 || p.Instances != 4 {
		t.Errorf("pooling counts = %d/%d, want 2/4", p.Workloads, p.Instances)
	}

	// durability: velero alone
	d := byID["durability"]
	if d.Workloads != 1 || d.Instances != 1 {
		t.Errorf("durability counts = %d/%d, want 1/1", d.Workloads, d.Instances)
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
		for _, b := range s.Blocks {
			total++
			if b.Namespace != "authentik" || b.Name != "authentik-server" {
				t.Errorf("unexpected block %+v", b)
			}
		}
	}
	if total != 1 {
		t.Fatalf("same workload split across %d blocks, want exactly 1", total)
	}
}
