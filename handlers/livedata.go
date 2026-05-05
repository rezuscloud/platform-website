package handlers

import "github.com/rezuscloud/platform-website/obs"

// defaultLiveData returns the real cluster topology for the live section.
// When the SigNoz client is wired in, this will be replaced by Client.Fetch().
func defaultLiveData() obs.LiveData {
	return obs.LiveData{
		Nodes: []obs.Node{
			{
				Name:   "talos-oci-cp-0",
				Tier:   "oci-cloud",
				Status: "Ready",
				CPU:    "12%",
				Mem:    "4.2 GiB",
				Pods: []obs.Pod{
					{Name: "cilium-operator-6d8f9c7b4-x2k1p", Namespace: "kube-system", Status: "Running", Restarts: 0},
					{Name: "coredns-7db8d4f5b9-mn3vq", Namespace: "kube-system", Status: "Running", Restarts: 0},
				},
			},
			{
				Name:   "talos-oci-cp-1",
				Tier:   "oci-cloud",
				Status: "Ready",
				CPU:    "8%",
				Mem:    "3.1 GiB",
				Pods: []obs.Pod{
					{Name: "etcd-talos-oci-cp-1", Namespace: "kube-system", Status: "Running", Restarts: 0},
					{Name: "kube-apiserver-talos-oci-cp-1", Namespace: "kube-system", Status: "Running", Restarts: 0},
				},
			},
			{
				Name:   "talosedge-genmachiche-flowing-bluejay",
				Tier:   "edge",
				Status: "Ready",
				CPU:    "34%",
				Mem:    "5.8 GiB",
				Pods: []obs.Pod{
					{Name: "platform-website-6f8d6fd5fc-4rgj9", Namespace: "platform-website-pr57", Status: "Running", Restarts: 0},
					{Name: "signoz-otel-collector-7c9d8f6b5-jk2mn", Namespace: "signoz", Status: "Running", Restarts: 0},
					{Name: "signoz-query-service-0", Namespace: "signoz", Status: "Running", Restarts: 0},
				},
			},
		},
	}
}
