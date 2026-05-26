package obs

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var meterProvider *sdkmetric.MeterProvider
var tracerProvider *sdktrace.TracerProvider

func InitTelemetry() (metric.MeterProvider, trace.TracerProvider) {
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

	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("Failed to create OTLP metric exporter: %v", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("Failed to create OTLP trace exporter: %v", err)
	}

	meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
			sdkmetric.WithInterval(15*time.Second),
		)),
	)

	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter,
			sdktrace.WithBatchTimeout(5*time.Second),
		),
		sdktrace.WithSampler(sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(0.1),
		)),
	)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if err := runtime.Start(runtime.WithMeterProvider(meterProvider)); err != nil {
		log.Fatalf("Failed to start OTel runtime instrumentation: %v", err)
	}

	return meterProvider, tracerProvider
}

func OTelFiberMiddleware(provider metric.MeterProvider, tp trace.TracerProvider) fiber.Handler {
	return otelfiber.Middleware(
		otelfiber.WithMeterProvider(provider),
		otelfiber.WithTracerProvider(tp),
	)
}

func ShutdownTelemetry(ctx context.Context) {
	if tracerProvider != nil {
		_ = tracerProvider.Shutdown(ctx)
	}
	if meterProvider != nil {
		_ = meterProvider.Shutdown(ctx)
	}
}
