package main

import (
	"context"

	"github.com/couchbase/gocb/v2"
	sdk "github.com/wasmCloud/provider-sdk-go"
	wrpc "github.com/wrpc/wrpc/go"

	// Generated bindings
	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/exports/wasmcloud/couchbase/document"
	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/wasmcloud/couchbase/types"
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

// GetAllReplicas implements document.Handler.
func (h *Handler) GetAllReplicas(ctx__ context.Context, id string, options *document.DocumentGetAllReplicaOptions) (*wrpc.Result[[]*document.DocumentGetReplicaResult, types.DocumentError], error) {
	panic("unimplemented")
}

// GetAndLock implements document.Handler.
func (h *Handler) GetAndLock(ctx__ context.Context, id string, options *document.DocumentGetAndLockOptions) (*wrpc.Result[document.DocumentGetResult, types.DocumentError], error) {
	panic("unimplemented")
}

// GetAndTouch implements document.Handler.
func (h *Handler) GetAndTouch(ctx__ context.Context, id string, options *document.DocumentGetAndTouchOptions) (*wrpc.Result[document.DocumentGetResult, types.DocumentError], error) {
	panic("unimplemented")
}

// GetAnyRepliacs implements document.Handler.
func (h *Handler) GetAnyRepliacs(ctx__ context.Context, id string, options *document.DocumentGetAnyReplicaOptions) (*wrpc.Result[document.DocumentGetReplicaResult, types.DocumentError], error) {
	panic("unimplemented")
}

// Insert implements document.Handler.
func (h *Handler) Insert(ctx__ context.Context, id string, doc *types.Document, options *document.DocumentInsertOptions) (*wrpc.Result[types.MutationMetadata, types.DocumentError], error) {
	panic("unimplemented")
}

// Remove implements document.Handler.
func (h *Handler) Remove(ctx__ context.Context, id string, options *document.DocumentRemoveOptions) (*wrpc.Result[types.MutationMetadata, types.DocumentError], error) {
	panic("unimplemented")
}

// Replace implements document.Handler.
func (h *Handler) Replace(ctx__ context.Context, id string, doc *types.Document, options *document.DocumentReplaceOptions) (*wrpc.Result[types.MutationMetadata, types.DocumentError], error) {
	panic("unimplemented")
}

// Touch implements document.Handler.
func (h *Handler) Touch(ctx__ context.Context, id string, options *document.DocumentTouchOptions) (*wrpc.Result[types.MutationMetadata, types.DocumentError], error) {
	panic("unimplemented")
}

// Unlock implements document.Handler.
func (h *Handler) Unlock(ctx__ context.Context, id string, options *document.DocumentUnlockOptions) (*wrpc.Result[struct{}, types.DocumentError], error) {
	panic("unimplemented")
}

// Upsert implements document.Handler.
func (h *Handler) Upsert(ctx__ context.Context, id string, doc *types.Document, options *document.DocumentUpsertOptions) (*wrpc.Result[types.MutationMetadata, types.DocumentError], error) {
	panic("unimplemented")
}

func (h *Handler) Get(context.Context, string, *document.DocumentGetOptions) (*wrpc.Result[document.DocumentGetResult, types.DocumentError], error) {
	return nil, nil
}
