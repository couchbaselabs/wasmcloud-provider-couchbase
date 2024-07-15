//go:generate wit-bindgen-wrpc go --out-dir bindings --package github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings wit

package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	server "github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings"
	"github.com/couchbase/gocb/v2"
	"github.com/wasmCloud/provider-sdk-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		return err
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// Initialize the provider with callbacks to track linked components
	providerHandler := Handler{
		linkedFrom:         make(map[string]map[string]string),
		clusterConnections: make(map[string]*gocb.Collection),
	}

	p, err := provider.New(
		provider.TargetLinkPut(func(link provider.InterfaceLinkDefinition) error {
			return handleNewTargetLink(&providerHandler, link)
		}),
		provider.TargetLinkDel(func(link provider.InterfaceLinkDefinition) error {
			return handleDelTargetLink(&providerHandler, link)
		}),
		provider.HealthCheck(func() string {
			return handleHealthCheck(&providerHandler)
		}),
		provider.Shutdown(func() error {
			return handleShutdown(&providerHandler)
		}),
	)
	if err != nil {
		return err
	}

	// Store the provider for use in the handlers
	providerHandler.WasmcloudProvider = p

	// Setup two channels to await RPC and control interface operations
	providerCh := make(chan error, 1)
	signalCh := make(chan os.Signal, 1)

	// Handle RPC operations
	stopFunc, err := server.Serve(p.RPCClient, &providerHandler, &providerHandler)
	if err != nil {
		p.Shutdown()
		return err
	}

	// Handle control interface operations
	go func() {
		err := p.Start()
		providerCh <- err
	}()

	// Shutdown on SIGINT
	signal.Notify(signalCh, syscall.SIGINT)

	// Run provider until either a shutdown is requested or a SIGINT is received
	select {
	case err = <-providerCh:
		stopFunc()
		return err
	case <-signalCh:
		p.Shutdown()
		stopFunc()
	}
	return nil
}

func handleNewTargetLink(handler *Handler, link provider.InterfaceLinkDefinition) error {
	handler.Logger.Info("Handling new target link", "link", link)
	handler.linkedFrom[link.SourceID] = link.TargetConfig
	err := ValidateCouchbaseConfig(link.TargetConfig)
	if err != nil {
		handler.Logger.Error("Invalid couchbase target config", "error", err)
		return err
	}
	handler.updateCouchbaseCluster(handler, link.SourceID, link.TargetConfig)
	return nil
}

func ValidateCouchbaseConfig(config map[string]string) error {
	if config["username"] == "" {
		return errors.New("username is required")
	}
	if config["password"] == "" {
		return errors.New("password is required")
	}
	if config["bucket"] == "" {
		return errors.New("bucket is required")
	}
	if config["host"] == "" {
		return errors.New("host is required")
	}
	return nil
}

func handleDelTargetLink(handler *Handler, link provider.InterfaceLinkDefinition) error {
	handler.Logger.Info("Handling del target link", "link", link)
	delete(handler.linkedFrom, link.Target)
	return nil
}

func handleHealthCheck(handler *Handler) string {
	handler.Logger.Info("Handling health check")
	return "provider healthy"
}

func handleShutdown(handler *Handler) error {
	handler.Logger.Info("Handling shutdown")
	// clear(handler.linkedFrom)
	return nil
}

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
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

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider() (*trace.TracerProvider, error) {
	traceExporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
	)
	return traceProvider, nil
}
