package main

import (
	"context"

	"github.com/couchbase/gocb/v2"
	sdk "github.com/wasmCloud/provider-sdk-go"
	wrpc "github.com/wrpc/wrpc/go"

	// Generated bindings
	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/exports/wrpc/keyvalue/atomics"
	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/exports/wrpc/keyvalue/store"
)

var (
	errNoSuchStore     = store.NewError_NoSuchStore()
	errInvalidDataType = store.NewError_Other("invalid data type stored in map")
)

// This provider `Handler` stores a global collection for querying.
// TODO(#): Support storing connections per linked component
type Handler struct {
	// The provider instance
	*sdk.WasmcloudProvider
	// The couchbase collection
	collection *gocb.Collection
	// All components linked to this provider and their config.
	linkedFrom map[string]map[string]string
}

// Implementation of wasi:keyvalue/store

func (h *Handler) Get(ctx context.Context, bucket string, key string) (*wrpc.Result[[]uint8, store.Error], error) {
	h.Logger.Debug("received request to get value", "key", key)
	res, err := h.collection.Get(key, &gocb.GetOptions{Transcoder: gocb.NewRawJSONTranscoder()})
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

func (h *Handler) Set(ctx context.Context, bucket string, key string, value []uint8) (*wrpc.Result[struct{}, store.Error], error) {
	h.Logger.Debug("received request to set value", "key", key)
	_, err := h.collection.Upsert(key, &value, &gocb.UpsertOptions{Transcoder: gocb.NewRawJSONTranscoder()})
	if err != nil {
		h.Logger.Error("unable to store value", "key", key, "error", err)
		return wrpc.Err[struct{}](*errInvalidDataType), err
	}
	return wrpc.Ok[store.Error](struct{}{}), nil
}

func (h *Handler) Delete(ctx context.Context, bucket string, key string) (*wrpc.Result[struct{}, store.Error], error) {
	h.Logger.Debug("received request to delete value", "key", key)
	_, err := h.collection.Remove(key, nil)
	if err != nil {
		h.Logger.Error("unable to remove value", "key", key, "error", err)
		return wrpc.Err[struct{}](*errNoSuchStore), err
	}
	return wrpc.Ok[store.Error](struct{}{}), nil
}

func (h *Handler) Exists(ctx context.Context, bucket string, key string) (*wrpc.Result[bool, store.Error], error) {
	h.Logger.Debug("received request to check value existence", "key", key)
	res, err := h.collection.Exists(key, nil)
	if err != nil {
		h.Logger.Error("unable to check existence of value", "key", key, "error", err)
		return wrpc.Err[bool](*errNoSuchStore), err
	}
	return wrpc.Ok[store.Error](res.Exists()), nil
}

func (h *Handler) ListKeys(ctx context.Context, bucket string, cursor *uint64) (*wrpc.Result[store.KeyResponse, store.Error], error) {
	h.Logger.Warn("received request to list keys")
	return wrpc.Err[store.KeyResponse](*store.NewError_Other("list-keys operation not supported")), nil
}

// Implementation of wasi:keyvalue/atomics
func (h *Handler) Increment(ctx context.Context, bucket string, key string, delta uint64) (*wrpc.Result[uint64, atomics.Error], error) {
	h.Logger.Debug("received request to increment key by delta", "key", key, "delta", delta)
	res, err := h.collection.Binary().Increment(key, &gocb.IncrementOptions{Initial: int64(delta), Delta: delta})
	if err != nil {
		h.Logger.Error("unable to increment value at key", "key", key, "error", err)
		return wrpc.Err[uint64](*errInvalidDataType), err
	}

	return wrpc.Ok[atomics.Error](res.Content()), nil
}
