package main

import (
	"context"
	"errors"
	"time"

	"github.com/couchbase/gocb/v2"
	sdk "github.com/wasmCloud/provider-sdk-go"
	wrpc "github.com/wrpc/wrpc/go"
	wrpcnats "github.com/wrpc/wrpc/go/nats"

	// Generated bindings
	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/exports/wrpc/keyvalue/atomics"
	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/exports/wrpc/keyvalue/store"
)

var (
	errNoSuchStore     = store.NewErrorNoSuchStore()
	errInvalidDataType = store.NewErrorOther("invalid data type stored in map")
)

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

// Implementation of wasi:keyvalue/store

func (h *Handler) Get(ctx context.Context, bucket string, key string) (*wrpc.Result[[]uint8, store.Error], error) {
	h.Logger.Debug("received request to get value", "key", key)
	collection, err := h.getCollectionFromContext(ctx)
	res, err := collection.Get(key, &gocb.GetOptions{Transcoder: gocb.NewRawJSONTranscoder()})
	if err != nil {
		h.Logger.Error("unable to get value in store", "key", key, "error", err)
		return wrpc.Err[[]uint8](*errNoSuchStore), err
	}

	var response []uint8
	err = res.Content(&response)
	if err != nil {
		h.Logger.Error("unable to decode content as bytes", "key", key, "error", err)
		return wrpc.Err[[]uint8](*errInvalidDataType), err
	}
	return wrpc.Ok[store.Error](response), nil
}

func (h *Handler) getCollectionFromContext(ctx context.Context) (*gocb.Collection, error) {
	header, ok := wrpcnats.HeaderFromContext(ctx)
	if !ok {
		h.Logger.Warn("Received request from unknown origin")
		return nil, errors.New("Error fetching header from wrpc context")
	}
	// Only allow requests from a linked component
	sourceId := header.Get("source-id")
	if h.linkedFrom[sourceId] == nil {
		h.Logger.Warn("Received request from unlinked source", "sourceId", sourceId)
		return nil, errors.New("Received request from unlinked source")
	}
	return h.clusterConnections[sourceId], nil
}

func (h *Handler) Set(ctx context.Context, bucket string, key string, value []uint8) (*wrpc.Result[struct{}, store.Error], error) {
	h.Logger.Debug("received request to set value", "key", key)
	collection, err := h.getCollectionFromContext(ctx)
	_, err = collection.Upsert(key, &value, &gocb.UpsertOptions{Transcoder: gocb.NewRawJSONTranscoder()})
	if err != nil {
		h.Logger.Error("unable to store value", "key", key, "error", err)
		return wrpc.Err[struct{}](*errInvalidDataType), err
	}
	return wrpc.Ok[store.Error](struct{}{}), nil
}

func (h *Handler) Delete(ctx context.Context, bucket string, key string) (*wrpc.Result[struct{}, store.Error], error) {
	h.Logger.Debug("received request to delete value", "key", key)
	collection, err := h.getCollectionFromContext(ctx)
	_, err = collection.Remove(key, nil)
	if err != nil {
		h.Logger.Error("unable to remove value", "key", key, "error", err)
		return wrpc.Err[struct{}](*errNoSuchStore), err
	}
	return wrpc.Ok[store.Error](struct{}{}), nil
}

func (h *Handler) Exists(ctx context.Context, bucket string, key string) (*wrpc.Result[bool, store.Error], error) {
	h.Logger.Debug("received request to check value existence", "key", key)
	collection, err := h.getCollectionFromContext(ctx)
	res, err := collection.Exists(key, nil)
	if err != nil {
		h.Logger.Error("unable to check existence of value", "key", key, "error", err)
		return wrpc.Err[bool](*errNoSuchStore), err
	}
	return wrpc.Ok[store.Error](res.Exists()), nil
}

func (h *Handler) ListKeys(ctx context.Context, bucket string, cursor *uint64) (*wrpc.Result[store.KeyResponse, store.Error], error) {
	h.Logger.Warn("received request to list keys")
	return wrpc.Err[store.KeyResponse](*store.NewErrorOther("list-keys operation not supported")), nil
}

// Implementation of wasi:keyvalue/atomics
func (h *Handler) Increment(ctx context.Context, bucket string, key string, delta uint64) (*wrpc.Result[uint64, atomics.Error], error) {
	h.Logger.Debug("received request to increment key by delta", "key", key, "delta", delta)
	collection, err := h.getCollectionFromContext(ctx)
	res, err := collection.Binary().Increment(key, &gocb.IncrementOptions{Initial: int64(delta), Delta: delta})
	if err != nil {
		h.Logger.Error("unable to increment value at key", "key", key, "error", err)
		return wrpc.Err[uint64](*errInvalidDataType), err
	}

	return wrpc.Ok[atomics.Error](res.Content()), nil
}

func (h *Handler) updateCouchbaseCluster(handler *Handler, sourceId string, config map[string]string) {
	// Connect to the cluster
	cluster, err := gocb.Connect(config["connectionString"], gocb.ClusterOptions{
		Username: config["username"],
		Password: config["password"],
	})
	if err != nil {
		handler.Logger.Error("unable to connect to couchbase cluster", "error", err)
		return
	}
	var collection *gocb.Collection
	bucketName := config["bucketName"]
	scopeName := config["scopeName"]
	collectionName := config["collectionName"]
	if collectionName != "" && scopeName != "" {
		collection = cluster.Bucket(bucketName).Scope(scopeName).Collection(collectionName)
	} else {
		collection = cluster.Bucket(bucketName).DefaultCollection()
	}

	bucket := cluster.Bucket(bucketName)
	if err = bucket.WaitUntilReady(5*time.Second, nil); err != nil {
		handler.Logger.Error("unable to connect to couchbase bucket", "error", err)
	}

	// Store the connection
	handler.clusterConnections[sourceId] = collection
}
