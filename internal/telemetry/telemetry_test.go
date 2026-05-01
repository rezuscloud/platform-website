package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInit_NoEndpoint(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	shutdown, err := Init(context.Background(), "test-service")

	assert.NoError(t, err)
	assert.NotNil(t, shutdown)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	assert.NoError(t, shutdown(shutdownCtx))
}

func TestInit_WithEndpoint_ReturnsShutdown(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")

	shutdown, err := Init(context.Background(), "test-service")

	assert.NoError(t, err)
	assert.NotNil(t, shutdown)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = shutdown(ctx)
}
