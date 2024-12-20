package provider

import (
	"context"
	"errors"
	"time"

	gocbt "github.com/couchbase/gocb-opentelemetry"
	"github.com/couchbase/gocb/v2"
	"go.opentelemetry.io/otel"
	sdk "go.wasmcloud.dev/provider"
	wrpcnats "wrpc.io/go/nats"
	// Generated bindings
)

const TRACER_NAME = "wasmcloud-provider-couchbase"

// This provider `Handler` stores a global collection for querying.
// TODO(#): Support storing connections per linked component
type Handler struct {
	// The provider instance
	*sdk.WasmcloudProvider
	// All components linked to this provider and their config.
	linkedFrom map[string]map[string]string

	// map that stores couchbase cluster connections
	clusterConnections map[string]*gocb.Collection
}

func NewLinkHandler() *Handler {
	return &Handler{
		linkedFrom:         make(map[string]map[string]string),
		clusterConnections: make(map[string]*gocb.Collection),
	}
}

// Provider handler functions
func (h *Handler) HandleNewTargetLink(link sdk.InterfaceLinkDefinition) error {
	h.Logger.Info("Handling new target link", "link", link)
	h.linkedFrom[link.SourceID] = link.TargetConfig
	couchbaseConnectionArgs, err := validateCouchbaseConfig(link.TargetConfig, link.TargetSecrets)
	if err != nil {
		h.Logger.Error("Invalid couchbase target config", "error", err)
		return err
	}
	h.updateCouchbaseCluster(link.SourceID, couchbaseConnectionArgs)
	return nil
}

func (h *Handler) HandleDelTargetLink(link sdk.InterfaceLinkDefinition) error {
	h.Logger.Info("Handling del target link", "link", link)
	delete(h.linkedFrom, link.Target)
	return nil
}

func (h *Handler) HandleHealthCheck() string {
	h.Logger.Debug("Handling health check")
	return "provider healthy"
}

func (h *Handler) HandleShutdown() error {
	h.Logger.Info("Handling shutdown")
	// clear(handler.linkedFrom)
	return nil
}

func (h *Handler) updateCouchbaseCluster(sourceId string, connectionArgs CouchbaseConnectionArgs) {
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
	var collection *gocb.Collection
	if connectionArgs.CollectionName != "" && connectionArgs.ScopeName != "" {
		collection = cluster.Bucket(connectionArgs.BucketName).Scope(connectionArgs.ScopeName).Collection(connectionArgs.CollectionName)
	} else {
		collection = cluster.Bucket(connectionArgs.BucketName).DefaultCollection()
	}

	bucket := cluster.Bucket(connectionArgs.BucketName)
	if err = bucket.WaitUntilReady(5*time.Second, nil); err != nil {
		h.Logger.Error("unable to connect to couchbase bucket", "error", err)
	}

	// Store the connection
	h.clusterConnections[sourceId] = collection
}

// Helper function to get the correct collection from the invocation context
func (h *Handler) GetCollectionFromContext(ctx context.Context) (*gocb.Collection, error) {
	header, ok := wrpcnats.HeaderFromContext(ctx)
	if !ok {
		h.Logger.Warn("Received request from unknown origin")
		return nil, errors.New("error fetching header from wrpc context")
	}
	// Only allow requests from a linked component
	sourceId := header.Get("source-id")
	if h.linkedFrom[sourceId] == nil {
		h.Logger.Warn("Received request from unlinked source", "sourceId", sourceId)
		return nil, errors.New("received request from unlinked source")
	}
	return h.clusterConnections[sourceId], nil
}
