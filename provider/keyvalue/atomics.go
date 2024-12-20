package keyvalue

import (
	"context"

	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/exports/wrpc/keyvalue/atomics"
	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/exports/wrpc/keyvalue/store"
	"github.com/couchbase/gocb/v2"
	wrpc "wrpc.io/go"
)

// ensure KeyvalueHandler implements atomics.Handler
var _ atomics.Handler = &KeyvalueHandler{}

// Increment implements atomics.Handler
func (h *KeyvalueHandler) Increment(ctx context.Context, bucket string, key string, delta uint64) (*wrpc.Result[uint64, store.Error], error) {
	collection, err := h.provider.GetCollectionFromContext(ctx)
	if err != nil {
		h.logger.Error("unable to get collection from context", "error", err)
		return wrpc.Err[uint64](*store.NewErrorNoSuchStore()), err
	}

	res, err := collection.Binary().Increment(key, &gocb.IncrementOptions{
		Initial: int64(delta),
		Delta:   delta,
	})
	if err != nil {
		h.logger.Error("unable to increment value at key", "error", err)
		return wrpc.Err[uint64](*store.NewErrorOther("invalid data type stored in map")), err
	}

	return wrpc.Ok[atomics.Error](res.Content()), nil
}
