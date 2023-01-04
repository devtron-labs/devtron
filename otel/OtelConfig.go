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
	"google.golang.org/grpc/credentials"
	"log"
)

type OtelConfig struct {
	OtelCollectorUrl string `env:"OTEL_COLLECTOR_URL" envDefault:""`
}

// Init configures an OpenTelemetry exporter and trace provider
func Init(serviceName string) *sdktrace.TracerProvider {
	//var collectorURL = "otel-collector.observability:4317"
	otelCfg := &OtelConfig{}
	err := env.Parse(otelCfg)
	if err != nil {
		log.Println("error occurred while parsing otel config", err)
		return nil
	}

	if otelCfg.OtelCollectorUrl == "" { // otel is not configured
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
		log.Println("error occurred while connecting to exporter", err)
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

	return traceProvider
}

type OtelSpan struct {
	reqContext       context.Context
	OverridenContext context.Context
	span             trace.Span
}

func (impl OtelSpan) End() {
	impl.span.End()
}

func StartSpan(serviceName, spanName string, ctx context.Context) OtelSpan {
	newCtx, span := otel.Tracer(serviceName).Start(ctx, spanName)
	otelSpan := OtelSpan{
		reqContext:       ctx,
		OverridenContext: newCtx,
		span:             span,
	}
	return otelSpan
}
