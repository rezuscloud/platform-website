package obs

// CategoryForNamespace maps a namespace to a platform category.
func CategoryForNamespace(ns string) string {
	switch ns {
	case "forgejo", "arc-systems":
		return "dev"
	case "flux-system", "vela-system", "external-dns", "cert-manager":
		return "deployment"
	case "kube-system", "platform-website", "dapr-system":
		return "runtime"
	case "signoz", "monitoring":
		return "observability"
	case "tikv-system", "juicefs-csi", "velero":
		return "data"
	default:
		return "runtime"
	}
}

// CategoryOrder defines the row group order.
var CategoryOrder = []string{"dev", "deployment", "runtime", "observability", "data"}

// CategoryLabel returns the display name for a category ID.
func CategoryLabel(id string) string {
	switch id {
	case "hosts":
		return "Hosts"
	case "dev":
		return "Development"
	case "deployment":
		return "Deployment"
	case "runtime":
		return "Runtime"
	case "observability":
		return "Observability"
	case "data":
		return "Data"
	default:
		return id
	}
}

// SortServices sorts services by category order then name.
func SortServices(services []Service) {
	catOrder := map[string]int{}
	for i, c := range CategoryOrder {
		catOrder[c] = i
	}
	for i := 0; i < len(services); i++ {
		for j := i + 1; j < len(services); j++ {
			ci := catOrder[services[i].Category]
			cj := catOrder[services[j].Category]
			if ci > cj || (ci == cj && services[i].Name > services[j].Name) {
				services[i], services[j] = services[j], services[i]
			}
		}
	}
}
