//go:generate wit-bindgen-wrpc go --out-dir bindings --package github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings wit

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	server "github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings"
	"github.com/couchbase/gocb/v2"
	"github.com/wasmCloud/provider-sdk-go"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Initialize the provider with callbacks to track linked components
	providerHandler := Handler{}
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

	// couchbase setup start
	bucketName := p.HostData().Config["bucketName"]
	connectionString := p.HostData().Config["connectionString"]
	username := p.HostData().Config["username"]
	password := p.HostData().Config["password"]
	scopeName := p.HostData().Config["scope"]
	collectionName := p.HostData().Config["collection"]

	cluster, err := gocb.Connect("couchbase://"+connectionString, gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: username,
			Password: password,
		},
	})
	if err != nil {
		return err
	}

	bucket := cluster.Bucket(bucketName)
	if err = bucket.WaitUntilReady(5*time.Second, nil); err != nil {
		return err
	}
	scope := bucket.Scope(scopeName)
	col := scope.Collection(collectionName)

	providerHandler.collection = col
	// couchbase setup end

	// Setup two channels to await RPC and control interface operations
	providerCh := make(chan error, 1)
	signalCh := make(chan os.Signal, 1)

	// Handle RPC operations
	stopFunc, err := server.Serve(p.RPCClient, &providerHandler)
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
	// handler.Logger.Info("Handling new target link", "link", link)
	// handler.linkedFrom[link.Target] = link.TargetConfig
	return nil
}

func handleDelTargetLink(handler *Handler, link provider.InterfaceLinkDefinition) error {
	// handler.Logger.Info("Handling del target link", "link", link)
	// delete(handler.linkedFrom, link.Target)
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
