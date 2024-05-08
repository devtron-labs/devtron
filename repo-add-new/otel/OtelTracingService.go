package otel

import (
	"context"
	"github.com/caarlos0/env/v6"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

const OTEL_ORCHESTRASTOR_SERVICE_NAME = "orchestrator"

type OtelTracingService interface {
	Init(serviceName string) *sdktrace.TracerProvider
	Shutdown()
}

type OtelTracingServiceImpl struct {
	logger        *zap.SugaredLogger
	traceProvider *sdktrace.TracerProvider
}

func NewOtelTracingServiceImpl(logger *zap.SugaredLogger) *OtelTracingServiceImpl {
	return &OtelTracingServiceImpl{logger: logger}
}

type OtelConfig struct {
	OtelCollectorUrl string `env:"OTEL_COLLECTOR_URL" envDefault:""`
}

// Init configures an OpenTelemetry exporter and trace provider
func (impl OtelTracingServiceImpl) Init(serviceName string) *sdktrace.TracerProvider {
	//var collectorURL = "otel-collector.observability:4317"
	otelCfg := &OtelConfig{}
	err := env.Parse(otelCfg)
	if err != nil {
		impl.logger.Errorw("error occurred while parsing otel config", "err", err)
		return nil
	}

	if otelCfg.OtelCollectorUrl == "" { // otel is not configured
		noopTracerProvider := trace.NewNoopTracerProvider()
		otel.SetTracerProvider(noopTracerProvider)
		return nil
	}

	secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")) // config can be passed to configure TLS
	secureOption = otlptracegrpc.WithInsecure()

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(otelCfg.OtelCollectorUrl),
		),
	)
	if err != nil {
		impl.logger.Errorw("error occurred while connecting to exporter", "err", err)
		return nil
	}

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String(serviceName))),
	)

	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	impl.logger.Infow("otel configured", "url", otelCfg.OtelCollectorUrl)
	impl.traceProvider = traceProvider
	return traceProvider
}

func (impl OtelTracingServiceImpl) Shutdown() {
	impl.logger.Info("shutting down trace")
	if impl.traceProvider == nil {
		impl.logger.Info("trace shutdown ignored as not enabled")
		return
	}
	if err := impl.traceProvider.Shutdown(context.Background()); err != nil {
		impl.logger.Errorw("Error shutting down tracer provider: ", "err", err)
	}
	impl.logger.Info("trace shutdown success")
}
