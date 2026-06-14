package obs

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// LoadSecretsFromDapr pulls optional config from the Dapr secret store.
// Currently only PROMETHEUS_URL is loaded (when the sidecar is present); the
// kube-prometheus-stack service is used by default otherwise.
func LoadSecretsFromDapr() {
	port := os.Getenv("DAPR_HTTP_PORT")
	if port == "" {
		return
	}

	client := &http.Client{Timeout: 3 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	healthReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:"+port+"/v1.0/healthz", nil)
	resp, err := client.Do(healthReq)
	if err != nil {
		log.Printf("dapr: sidecar health check failed: %v", err)
		return
	}
	resp.Body.Close()

	secretReq, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		"http://localhost:"+port+"/v1.0/secrets/kubernetes-secret-store/observability-config", nil)
	secretResp, err := client.Do(secretReq)
	if err != nil {
		log.Printf("dapr: secrets fetch failed: %v", err)
		return
	}
	defer secretResp.Body.Close()

	if secretResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(secretResp.Body)
		log.Printf("dapr: secrets fetch returned %d: %s", secretResp.StatusCode, string(body[:min(len(body), 200)]))
		return
	}

	var secrets map[string]string
	if err := json.NewDecoder(secretResp.Body).Decode(&secrets); err != nil {
		log.Printf("dapr: decode secrets: %v", err)
		return
	}

	if u, ok := secrets["PROMETHEUS_URL"]; ok && u != "" {
		os.Setenv("PROMETHEUS_URL", u)
	}
	log.Printf("dapr: loaded observability config from kubernetes-secret-store")
}
