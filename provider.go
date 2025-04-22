//go:generate wit-bindgen-wrpc go --out-dir bindings --package github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings wit

package main

import (
	"time"

	gocbt "github.com/couchbase/gocb-opentelemetry"
	"github.com/couchbase/gocb/v2"
	"go.opentelemetry.io/otel"
	"go.wasmcloud.dev/provider"
)

// The primary function for connecting a sourceId component to a Couchbase cluster
func (h *Handler) updateCouchbaseCluster(sourceId string, linkName string, connectionArgs CouchbaseConnectionArgs) {
	// Connect to the cluster
	cluster, err := gocb.Connect(connectionArgs.ConnectionString, gocb.ClusterOptions{
		Username: connectionArgs.Username,
		Password: connectionArgs.Password,
		Tracer:   gocbt.NewOpenTelemetryRequestTracer(otel.GetTracerProvider()),
	})
	if err != nil {
		h.Logger.Error("unable to connect to couchbase cluster", "error", err)
		return
	}

	bucket := cluster.Bucket(connectionArgs.BucketName)
	if err = bucket.WaitUntilReady(5*time.Second, nil); err != nil {
		h.Logger.Error("unable to connect to couchbase bucket", "error", err)
		return
	}

	var collection *gocb.Collection
	if connectionArgs.CollectionName != "" && connectionArgs.ScopeName != "" {
		collection = bucket.Scope(connectionArgs.ScopeName).Collection(connectionArgs.CollectionName)
	} else if connectionArgs.ScopeName != "" || connectionArgs.CollectionName != "" {
		h.Logger.Warn("scopeName and collectionName must be provided together, using default collection")
		collection = bucket.DefaultCollection()
	} else {
		collection = bucket.DefaultCollection()
	}

	// Store the connection
	if h.clusterConnections == nil {
		h.clusterConnections = make(map[string]map[string]*gocb.Collection)
	}
	if h.clusterConnections[sourceId] == nil {
		h.clusterConnections[sourceId] = make(map[string]*gocb.Collection)
	}

	h.clusterConnections[sourceId][linkName] = collection
}

// Provider handler functions
func (h *Handler) handleNewTargetLink(link provider.InterfaceLinkDefinition) error {
	h.Logger.Info("Handling new target link", "link", link)
	couchbaseConnectionArgs, err := validateCouchbaseConfig(link.TargetConfig, link.TargetSecrets)
	if err != nil {
		h.Logger.Error("Invalid couchbase target config", "error", err)
		return err
	}
	h.updateCouchbaseCluster(link.SourceID, link.Name, couchbaseConnectionArgs)
	return nil
}

func (h *Handler) handleDelTargetLink(link provider.InterfaceLinkDefinition) error {
	h.Logger.Info("Handling del target link", "link", link)
	if connections, exists := h.clusterConnections[link.SourceID]; exists {
		delete(connections, link.Name)
		if len(connections) == 0 {
			delete(h.clusterConnections, link.SourceID)
		}
	}
	return nil
}

func (h *Handler) handleHealthCheck() string {
	h.Logger.Debug("Handling health check")
	return "provider healthy"
}

func (h *Handler) handleShutdown() error {
	h.Logger.Info("Handling shutdown")
	clear(h.clusterConnections)
	return nil
}
