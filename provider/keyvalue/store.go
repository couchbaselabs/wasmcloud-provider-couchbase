package keyvalue

import (
	"context"

	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/exports/wrpc/keyvalue/store"
	"github.com/couchbase/gocb/v2"
	wrpc "wrpc.io/go"
)

// ensure KeyvalueHandler implements store.Handler
var _ store.Handler = &KeyvalueHandler{}

// Get implements store.Handler
func (h *KeyvalueHandler) Get(ctx context.Context, bucket string, key string) (*wrpc.Result[[]uint8, store.Error], error) {
	collection, err := h.provider.GetCollectionFromContext(ctx)
	if err != nil {
		h.logger.Error("unable to get collection from context", "error", err)
		return wrpc.Err[[]uint8](*store.NewErrorNoSuchStore()), err
	}

	res, err := collection.Get(key, &gocb.GetOptions{
		Transcoder: gocb.NewRawJSONTranscoder(),
	})
	if err != nil {
		h.logger.Error("unable to get value in store", "error", err)
		return wrpc.Err[[]uint8](*store.NewErrorNoSuchStore()), err
	}

	var response []uint8
	err = res.Content(&response)
	if err != nil {
		h.logger.Error("unable to decode content as bytes", "error", err)
		return wrpc.Err[[]uint8](*store.NewErrorOther("invalid data type stored in map")), err
	}
	return wrpc.Ok[store.Error](response), nil
}

// Set implements store.Handler
func (h *KeyvalueHandler) Set(ctx context.Context, bucket string, key string, value []uint8) (*wrpc.Result[struct{}, store.Error], error) {
	collection, err := h.provider.GetCollectionFromContext(ctx)
	if err != nil {
		h.logger.Error("unable to get collection from context", "error", err)
		return wrpc.Err[struct{}](*store.NewErrorNoSuchStore()), err
	}

	_, err = collection.Upsert(key, &value, &gocb.UpsertOptions{
		Transcoder: gocb.NewRawJSONTranscoder(),
	})
	if err != nil {
		h.logger.Error("unable to store value", "error", err)
		return wrpc.Err[struct{}](*store.NewErrorOther("invalid data type stored in map")), err
	}
	return wrpc.Ok[store.Error](struct{}{}), nil
}

// Delete implements store.Handler
func (h *KeyvalueHandler) Delete(ctx context.Context, bucket string, key string) (*wrpc.Result[struct{}, store.Error], error) {
	collection, err := h.provider.GetCollectionFromContext(ctx)
	if err != nil {
		h.logger.Error("unable to get collection from context", "error", err)
		return wrpc.Err[struct{}](*store.NewErrorNoSuchStore()), err
	}

	_, err = collection.Remove(key, &gocb.RemoveOptions{})
	if err != nil {
		h.logger.Error("unable to remove value", "error", err)
		return wrpc.Err[struct{}](*store.NewErrorNoSuchStore()), err
	}
	return wrpc.Ok[store.Error](struct{}{}), nil
}

// Exists implements store.Handler
func (h *KeyvalueHandler) Exists(ctx context.Context, bucket string, key string) (*wrpc.Result[bool, store.Error], error) {
	collection, err := h.provider.GetCollectionFromContext(ctx)
	if err != nil {
		h.logger.Error("unable to get collection from context", "error", err)
		return wrpc.Err[bool](*store.NewErrorNoSuchStore()), err
	}

	res, err := collection.Exists(key, &gocb.ExistsOptions{})
	if err != nil {
		h.logger.Error("unable to check existence of value", "error", err)
		return wrpc.Err[bool](*store.NewErrorNoSuchStore()), err
	}
	return wrpc.Ok[store.Error](res.Exists()), nil
}

// ListKeys implements store.Handler
func (h *KeyvalueHandler) ListKeys(ctx context.Context, bucket string, cursor *uint64) (*wrpc.Result[store.KeyResponse, store.Error], error) {
	h.logger.Warn("received request to list keys")
	return wrpc.Err[store.KeyResponse](*store.NewErrorOther("list-keys operation not supported")), nil
}
