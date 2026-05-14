package obs

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var promHandler http.Handler

func InitTelemetry() metric.MeterProvider {
	res, err := resource.New(
		context.Background(),
		resource.WithFromEnv(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("platform-website"),
			semconv.ServiceVersionKey.String(os.Getenv("APP_VERSION")),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create OTel resource: %v", err)
	}

	promExporter, err := prometheus.New()
	if err != nil {
		log.Fatalf("Failed to create Prometheus exporter: %v", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(promExporter),
	)

	if err := runtime.Start(runtime.WithMeterProvider(meterProvider)); err != nil {
		log.Fatalf("Failed to start OTel runtime instrumentation: %v", err)
	}

	promHandler = promhttp.Handler()

	return meterProvider
}

func MetricsHandler() http.Handler {
	return promHandler
}

func OTelFiberMiddleware(provider metric.MeterProvider) fiber.Handler {
	return otelfiber.Middleware(
		otelfiber.WithMeterProvider(provider),
	)
}

func ShutdownTelemetry(_ context.Context) {}
