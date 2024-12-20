package couchbase

import (
	"context"
	"errors"
	"time"

	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/exports/wasmcloud/couchbase/document"
	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/wasmcloud/couchbase/types"
	"github.com/couchbase/gocb/v2"
	wrpc "wrpc.io/go"
)

func Ok[T any](v T) *wrpc.Result[T, types.DocumentError] {
	return wrpc.Ok[types.DocumentError](v)
}

func Err[T any](e types.DocumentError) *wrpc.Result[T, types.DocumentError] {
	return wrpc.Err[T](e)
}

// ensure CouchbaseHandler implements document.Handler
var _ document.Handler = &CouchbaseHandler{}

// Get implements document.Handler
func (h *CouchbaseHandler) Get(ctx context.Context, id string, options *document.DocumentGetOptions) (*wrpc.Result[document.DocumentGetResult, types.DocumentError], error) {
	collection, err := h.provider.GetCollectionFromContext(ctx)
	if err != nil {
		h.logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	couchbaseResult, err := collection.Get(id, GetOptions(options))
	if err != nil {
		h.logger.Error("Error getting document", "error", err)
		return Err[document.DocumentGetResult](*types.NewDocumentErrorNotFound()), nil
	}
	documentResult, err := GetResult(couchbaseResult)
	if err != nil {
		h.logger.Error("Error getting document result", "error", err)
		return Err[document.DocumentGetResult](*types.NewDocumentErrorNotJson()), nil
	}
	return Ok(documentResult), nil
}

// GetAllReplicas implements document.Handler.
func (h *CouchbaseHandler) GetAllReplicas(ctx context.Context, id string, options *document.DocumentGetAllReplicaOptions) (*wrpc.Result[[]*document.DocumentGetReplicaResult, types.DocumentError], error) {
	collection, err := h.provider.GetCollectionFromContext(ctx)
	if err != nil {
		h.logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	res, err := collection.GetAllReplicas(id, GetAllReplicaOptions(options))
	if err != nil {
		h.logger.Error("Error fetching all replicas", "error", err)
		return nil, err
	}
	return Ok(GetAllReplicasResult(res)), nil
}

// GetAndLock implements document.Handler.
func (h *CouchbaseHandler) GetAndLock(ctx context.Context, id string, options *document.DocumentGetAndLockOptions) (*wrpc.Result[document.DocumentGetResult, types.DocumentError], error) {
	collection, err := h.provider.GetCollectionFromContext(ctx)
	if err != nil {
		h.logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	couchbaseResult, err := collection.GetAndLock(id, time.Duration(options.LockTime), GetAndLockOptions(options))
	if err != nil {
		h.logger.Error("Error getting and locking document", "error", err)
		return nil, err
	}
	documentResult, err := GetResult(couchbaseResult)
	if err != nil {
		h.logger.Error("Error getting document result", "error", err)
		return nil, err
	}
	return Ok(documentResult), nil
}

// GetAndTouch implements document.Handler.
func (h *CouchbaseHandler) GetAndTouch(ctx context.Context, id string, options *document.DocumentGetAndTouchOptions) (*wrpc.Result[document.DocumentGetResult, types.DocumentError], error) {
	collection, err := h.provider.GetCollectionFromContext(ctx)
	if err != nil {
		h.logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	couchbaseResult, err := collection.GetAndTouch(id, time.Duration(options.ExpiresIn), GetAndTouchOptions(options))
	if err != nil {
		h.logger.Error("Error getting and touching document", "error", err)
		return nil, err
	}
	documentResult, err := GetResult(couchbaseResult)
	if err != nil {
		h.logger.Error("Error getting document result", "error", err)
		return nil, err
	}
	return Ok(documentResult), nil
}

// GetAnyRepliacs implements document.Handler.
func (h *CouchbaseHandler) GetAnyRepliacs(ctx context.Context, id string, options *document.DocumentGetAnyReplicaOptions) (*wrpc.Result[document.DocumentGetReplicaResult, types.DocumentError], error) {
	collection, err := h.provider.GetCollectionFromContext(ctx)
	if err != nil {
		h.logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	result, err := collection.GetAnyReplica(id, GetAnyReplicaOptions(options))
	if err != nil {
		h.logger.Error("Error getting any replica", "error", err)
		return nil, err
	}
	return Ok(GetReplicaResult(result)), nil
}

// Insert implements document.Handler.
func (h *CouchbaseHandler) Insert(ctx context.Context, id string, doc *types.Document, options *document.DocumentInsertOptions) (*wrpc.Result[types.MutationMetadata, types.DocumentError], error) {
	collection, err := h.provider.GetCollectionFromContext(ctx)
	if err != nil {
		h.logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	docToInsert, ok := doc.GetRaw()
	if !ok {
		h.logger.Error("Error getting raw document", "doc", doc)
		return Err[types.MutationMetadata](*types.NewDocumentErrorNotJson()), nil
	}
	result, err := collection.Insert(id, docToInsert, InsertOptions(options))
	if err != nil {
		h.logger.Error("Error inserting document", "error", err)
		if errors.Is(err, gocb.ErrDocumentExists) {
			return Err[types.MutationMetadata](*types.NewDocumentErrorAlreadyExists()), nil
		} else {
			return Err[types.MutationMetadata](*types.NewDocumentErrorInvalidValue()), nil
		}
	}
	return Ok(MutationMetadata(result)), nil
}

// Remove implements document.Handler.
func (h *CouchbaseHandler) Remove(ctx context.Context, id string, options *document.DocumentRemoveOptions) (*wrpc.Result[types.MutationMetadata, types.DocumentError], error) {
	collection, err := h.provider.GetCollectionFromContext(ctx)
	if err != nil {
		h.logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	result, err := collection.Remove(id, RemoveOptions(options))
	if err != nil {
		h.logger.Error("Error removing document", "error", err)
		return nil, err
	}
	return Ok(MutationMetadata(result)), nil
}

// Replace implements document.Handler.
func (h *CouchbaseHandler) Replace(ctx context.Context, id string, doc *types.Document, options *document.DocumentReplaceOptions) (*wrpc.Result[types.MutationMetadata, types.DocumentError], error) {
	collection, err := h.provider.GetCollectionFromContext(ctx)
	if err != nil {
		h.logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}

	replacement, ok := doc.GetRaw()
	if !ok {
		h.logger.Error("Error getting raw document", "doc", doc)
		return Err[types.MutationMetadata](*types.NewDocumentErrorNotJson()), nil
	}

	result, err := collection.Replace(id, replacement, ReplaceOptions(options))
	if err != nil {
		h.logger.Error("Error replacing document", "error", err)
		return nil, err
	}
	return Ok(MutationMetadata(result)), nil
}

// Touch implements document.Handler.
func (h *CouchbaseHandler) Touch(ctx context.Context, id string, options *document.DocumentTouchOptions) (*wrpc.Result[types.MutationMetadata, types.DocumentError], error) {
	collection, err := h.provider.GetCollectionFromContext(ctx)
	if err != nil {
		h.logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	result, err := collection.Touch(id, time.Duration(options.ExpiresIn), TouchOptions(options))
	if err != nil {
		h.logger.Error("Error touching document", "error", err)
		return nil, err
	}
	return Ok(MutationMetadata(result)), nil
}

// Unlock implements document.Handler.
func (h *CouchbaseHandler) Unlock(ctx context.Context, id string, options *document.DocumentUnlockOptions) (*wrpc.Result[struct{}, types.DocumentError], error) {
	collection, err := h.provider.GetCollectionFromContext(ctx)
	if err != nil {
		h.logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	err = collection.Unlock(id, gocb.Cas(options.Cas), UnlockOptions(options))
	if err != nil {
		h.logger.Error("Error unlocking document", "error", err)
		return nil, err
	}
	return Ok(struct{}{}), nil
}

// Upsert implements document.Handler.
func (h *CouchbaseHandler) Upsert(ctx context.Context, id string, doc *types.Document, options *document.DocumentUpsertOptions) (*wrpc.Result[types.MutationMetadata, types.DocumentError], error) {
	collection, err := h.provider.GetCollectionFromContext(ctx)
	if err != nil {
		h.logger.Error("Error fetching collection from context", "error", err)
		return nil, err
	}
	raw, ok := doc.GetRaw()
	if !ok {
		h.logger.Error("Error getting raw document", "doc", doc)
		return Err[types.MutationMetadata](*types.NewDocumentErrorNotJson()), nil
	}
	result, err := collection.Upsert(id, raw, UpsertOptions(options))
	if err != nil {
		h.logger.Error("Error upserting document", "error", err)
		return Err[types.MutationMetadata](*types.NewDocumentErrorNotJson()), nil
	}
	return Ok(MutationMetadata(result)), nil
}
