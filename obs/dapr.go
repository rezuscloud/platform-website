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
		"http://localhost:"+port+"/v1.0/secrets/kubernetes-secret-store/signoz-api-credentials", nil)
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

	if u, ok := secrets["SIGNOZ_URL"]; ok && u != "" {
		os.Setenv("SIGNOZ_URL", u)
	}
	if k, ok := secrets["SIGNOZ_API_KEY"]; ok && k != "" {
		os.Setenv("SIGNOZ_API_KEY", k)
	}
	log.Printf("dapr: loaded SIGNOZ credentials from kubernetes-secret-store")
}
