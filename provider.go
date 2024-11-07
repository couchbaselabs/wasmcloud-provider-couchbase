package main

import (
	"context"
	"errors"
	"time"

	"github.com/couchbase/gocb/v2"
	sdk "go.wasmcloud.dev/provider"
	wrpc "wrpc.io/go"
	wrpcnats "wrpc.io/go/nats"

	// Generated bindings
	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/exports/wasmcloud/couchbase/document"
	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/wasmcloud/couchbase/types"
)

const TRACER_NAME = "wasmcloud-provider-couchbase"

func Ok[T any](v T) *wrpc.Result[T, types.DocumentError] {
	return wrpc.Ok[types.DocumentError](v)
}

func Err[T any](e types.DocumentError) *wrpc.Result[T, types.DocumentError] {
	return wrpc.Err[T](e)
}

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

func (h *Handler) Get(ctx context.Context, id string, options *document.DocumentGetOptions) (*wrpc.Result[document.DocumentGetResult, types.DocumentError], error) {
	collection, err := h.getCollectionFromContext(ctx)
	if err != nil {
		h.Logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	couchbaseResult, err := collection.Get(id, GetOptions(options))
	if err != nil {
		h.Logger.Error("Error getting document", "error", err)
		return Err[document.DocumentGetResult](*types.NewDocumentErrorNotFound()), nil
	}
	documentResult, err := GetResult(couchbaseResult)
	if err != nil {
		h.Logger.Error("Error getting document result", "error", err)
		return Err[document.DocumentGetResult](*types.NewDocumentErrorNotJson()), nil
	}
	return Ok(documentResult), nil
}

// GetAllReplicas implements document.Handler.
func (h *Handler) GetAllReplicas(ctx context.Context, id string, options *document.DocumentGetAllReplicaOptions) (*wrpc.Result[[]*document.DocumentGetReplicaResult, types.DocumentError], error) {
	collection, err := h.getCollectionFromContext(ctx)
	if err != nil {
		h.Logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	res, err := collection.GetAllReplicas(id, GetAllReplicaOptions(options))
	if err != nil {
		h.Logger.Error("Error fetching all replicas", "error", err)
		return nil, err
	}
	return Ok(GetAllReplicasResult(res)), nil
}

// GetAndLock implements document.Handler.
func (h *Handler) GetAndLock(ctx context.Context, id string, options *document.DocumentGetAndLockOptions) (*wrpc.Result[document.DocumentGetResult, types.DocumentError], error) {
	collection, err := h.getCollectionFromContext(ctx)
	if err != nil {
		h.Logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	couchbaseResult, err := collection.GetAndLock(id, time.Duration(options.LockTime), GetAndLockOptions(options))
	if err != nil {
		h.Logger.Error("Error getting and locking document", "error", err)
		return nil, err
	}
	documentResult, err := GetResult(couchbaseResult)
	if err != nil {
		h.Logger.Error("Error getting document result", "error", err)
		return nil, err
	}
	return Ok(documentResult), nil
}

// GetAndTouch implements document.Handler.
func (h *Handler) GetAndTouch(ctx context.Context, id string, options *document.DocumentGetAndTouchOptions) (*wrpc.Result[document.DocumentGetResult, types.DocumentError], error) {
	collection, err := h.getCollectionFromContext(ctx)
	if err != nil {
		h.Logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	couchbaseResult, err := collection.GetAndTouch(id, time.Duration(options.ExpiresIn), GetAndTouchOptions(options))
	if err != nil {
		h.Logger.Error("Error getting and touching document", "error", err)
		return nil, err
	}
	documentResult, err := GetResult(couchbaseResult)
	if err != nil {
		h.Logger.Error("Error getting document result", "error", err)
		return nil, err
	}
	return Ok(documentResult), nil
}

// GetAnyRepliacs implements document.Handler.
func (h *Handler) GetAnyRepliacs(ctx context.Context, id string, options *document.DocumentGetAnyReplicaOptions) (*wrpc.Result[document.DocumentGetReplicaResult, types.DocumentError], error) {
	collection, err := h.getCollectionFromContext(ctx)
	if err != nil {
		h.Logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	result, err := collection.GetAnyReplica(id, GetAnyReplicaOptions(options))
	if err != nil {
		h.Logger.Error("Error getting any replica", "error", err)
		return nil, err
	}
	return Ok(GetReplicaResult(result)), nil
}

// Insert implements document.Handler.
func (h *Handler) Insert(ctx context.Context, id string, doc *types.Document, options *document.DocumentInsertOptions) (*wrpc.Result[types.MutationMetadata, types.DocumentError], error) {
	collection, err := h.getCollectionFromContext(ctx)
	if err != nil {
		h.Logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	docToInsert, ok := doc.GetRaw()
	if !ok {
		h.Logger.Error("Error getting raw document", "doc", doc)
		return Err[types.MutationMetadata](*types.NewDocumentErrorNotJson()), nil
	}
	result, err := collection.Insert(id, docToInsert, InsertOptions(options))
	if err != nil {
		h.Logger.Error("Error inserting document", "error", err)
		if errors.Is(err, gocb.ErrDocumentExists) {
			return Err[types.MutationMetadata](*types.NewDocumentErrorAlreadyExists()), nil
		} else {
			return Err[types.MutationMetadata](*types.NewDocumentErrorInvalidValue()), nil
		}
	}
	return Ok(MutationMetadata(result)), nil
}

// Remove implements document.Handler.
func (h *Handler) Remove(ctx context.Context, id string, options *document.DocumentRemoveOptions) (*wrpc.Result[types.MutationMetadata, types.DocumentError], error) {
	collection, err := h.getCollectionFromContext(ctx)
	if err != nil {
		h.Logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	result, err := collection.Remove(id, RemoveOptions(options))
	if err != nil {
		h.Logger.Error("Error removing document", "error", err)
		return nil, err
	}
	return Ok(MutationMetadata(result)), nil
}

// Replace implements document.Handler.
func (h *Handler) Replace(ctx context.Context, id string, doc *types.Document, options *document.DocumentReplaceOptions) (*wrpc.Result[types.MutationMetadata, types.DocumentError], error) {
	collection, err := h.getCollectionFromContext(ctx)
	if err != nil {
		h.Logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}

	replacement, ok := doc.GetRaw()
	if !ok {
		h.Logger.Error("Error getting raw document", "doc", doc)
		return Err[types.MutationMetadata](*types.NewDocumentErrorNotJson()), nil
	}

	result, err := collection.Replace(id, replacement, ReplaceOptions(options))
	if err != nil {
		h.Logger.Error("Error replacing document", "error", err)
		return nil, err
	}
	return Ok(MutationMetadata(result)), nil
}

// Touch implements document.Handler.
func (h *Handler) Touch(ctx context.Context, id string, options *document.DocumentTouchOptions) (*wrpc.Result[types.MutationMetadata, types.DocumentError], error) {
	collection, err := h.getCollectionFromContext(ctx)
	if err != nil {
		h.Logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	result, err := collection.Touch(id, time.Duration(options.ExpiresIn), TouchOptions(options))
	if err != nil {
		h.Logger.Error("Error touching document", "error", err)
		return nil, err
	}
	return Ok(MutationMetadata(result)), nil
}

// Unlock implements document.Handler.
func (h *Handler) Unlock(ctx context.Context, id string, options *document.DocumentUnlockOptions) (*wrpc.Result[struct{}, types.DocumentError], error) {
	collection, err := h.getCollectionFromContext(ctx)
	if err != nil {
		h.Logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	err = collection.Unlock(id, gocb.Cas(options.Cas), UnlockOptions(options))
	if err != nil {
		h.Logger.Error("Error unlocking document", "error", err)
		return nil, err
	}
	return Ok(struct{}{}), nil
}

// Upsert implements document.Handler.
func (h *Handler) Upsert(ctx context.Context, id string, doc *types.Document, options *document.DocumentUpsertOptions) (*wrpc.Result[types.MutationMetadata, types.DocumentError], error) {
	collection, err := h.getCollectionFromContext(ctx)
	if err != nil {
		h.Logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	raw, ok := doc.GetRaw()
	if !ok {
		h.Logger.Error("Error getting raw document", "doc", doc)
		return Err[types.MutationMetadata](*types.NewDocumentErrorNotJson()), nil
	}
	result, err := collection.Upsert(id, raw, UpsertOptions(options))
	if err != nil {
		h.Logger.Error("Error upserting document", "error", err)
		return Err[types.MutationMetadata](*types.NewDocumentErrorNotJson()), nil
	}
	return Ok(MutationMetadata(result)), nil
}

// Helper function to get the correct collection from the invocation context
func (h *Handler) getCollectionFromContext(ctx context.Context) (*gocb.Collection, error) {
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
