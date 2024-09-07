package main

import (
	"context"
	"errors"
	"time"

	wrpcnats "github.com/wrpc/wrpc/go/nats"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func setupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTraceProvider()
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)
	traceProvider := otel.GetTracerProvider()
	tracer = traceProvider.Tracer(TRACER_NAME)

	return
}

func newExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	exporter, err := otlptrace.New(
		ctx,
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint("localhost:4318"),
			otlptracehttp.WithInsecure(),
		),
	)
	return exporter, err
}

func newTraceProvider() (*trace.TracerProvider, error) {
	traceExporter, err := newExporter(context.Background())
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("couchbase-provider"),
		)))
	return traceProvider, nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// extractTracerHeaderContext extracts the trace context from the wrpc headers.
func extractTracerHeaderContext(ctx context.Context) context.Context {
	headers, ok := wrpcnats.HeaderFromContext(ctx)
	if !ok {
		return ctx
	}

	pr := propagation.MapCarrier{}
	pr.Set("traceparent", headers.Get("traceparent"))
	pr.Set("tracestate", headers.Get("tracestate"))
	pr.Set("baggage", headers.Get("baggage"))

	return otel.GetTextMapPropagator().Extract(ctx, pr)
}
