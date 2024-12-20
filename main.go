//go:generate wit-bindgen-wrpc go --out-dir bindings --package github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings wit

package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	wrpc "github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings"
	"github.com/couchbase-examples/wasmcloud-provider-couchbase/provider"
	"github.com/couchbase-examples/wasmcloud-provider-couchbase/provider/couchbase"
	"github.com/couchbase-examples/wasmcloud-provider-couchbase/provider/keyvalue"
	sdk "go.wasmcloud.dev/provider"
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
	linkHandler := provider.NewLinkHandler()

	p, err := sdk.New(
		sdk.TargetLinkPut(linkHandler.HandleNewTargetLink),
		sdk.TargetLinkDel(linkHandler.HandleDelTargetLink),
		sdk.HealthCheck(linkHandler.HandleHealthCheck),
		sdk.Shutdown(linkHandler.HandleShutdown),
	)
	if err != nil {
		return err
	}

	// Store the provider for use in the handlers
	linkHandler.WasmcloudProvider = p

	// Setup two channels to await RPC and control interface operations
	providerCh := make(chan error, 1)
	signalCh := make(chan os.Signal, 1)

	// Handle RPC operations
	stopFunc, err := wrpc.Serve(
		p.RPCClient,
		keyvalue.New(linkHandler),
		keyvalue.New(linkHandler),
		couchbase.New(linkHandler),
	)
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
