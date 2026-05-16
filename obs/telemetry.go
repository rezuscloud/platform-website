package obs

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var meterProvider *sdkmetric.MeterProvider

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

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "signoz-otel-collector.signoz.svc.cluster.local:4317"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("Failed to create OTLP metric exporter: %v", err)
	}

	meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(15*time.Second),
		)),
	)

	if err := runtime.Start(runtime.WithMeterProvider(meterProvider)); err != nil {
		log.Fatalf("Failed to start OTel runtime instrumentation: %v", err)
	}

	return meterProvider
}

func OTelFiberMiddleware(provider metric.MeterProvider) fiber.Handler {
	return otelfiber.Middleware(
		otelfiber.WithMeterProvider(provider),
	)
}

func ShutdownTelemetry(ctx context.Context) {
	if meterProvider != nil {
		_ = meterProvider.Shutdown(ctx)
	}
}
