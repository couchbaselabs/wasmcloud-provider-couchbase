package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/exports/wasmcloud/couchbase/document"
	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/wasmcloud/couchbase/types"
	"github.com/google/uuid"
	wrpc "wrpc.io/go"
)

type asyncResult[T any] struct {
	ready  bool
	result T
}

func (a *asyncResult[T]) HandleAsync(f func() T) {
	go func() {
		a.result = f()
		a.ready = true
	}()
}

func (h *Handler) InsertAsync(ctx context.Context, id string, doc *types.Document, options *document.DocumentInsertOptions) (wrpc.Own[document.InsertResultAsync], error) {
	h.Logger.Info("Handling InsertAsync")
	res := &asyncResult[*wrpc.Result[document.MutationMetadata, document.DocumentError]]{ready: false}
	asyncKey := fmt.Sprintf("insert.%s.%s", id, uuid.NewString())
	h.asyncResult[asyncKey] = res

	res.HandleAsync(func() *wrpc.Result[document.MutationMetadata, document.DocumentError] {
		res, err := h.Insert(ctx, id, doc, options)
		if err != nil {
			h.Logger.Error("Error InsertAsync failed", "error", err)
			return Err[types.MutationMetadata](*types.NewDocumentErrorNotFound())
		}
		return res
	})
	return wrpc.Own[document.InsertResultAsync](asyncKey), nil
}

func (h *Handler) InsertResultAsync_Ready(ctx__ context.Context, self wrpc.Borrow[document.InsertResultAsync]) (bool, error) {
	h.Logger.Info("InsertResultAsync Ready", "self", string(self))
	asyncKey := string(self)

	resVal, ok := h.asyncResult[asyncKey]
	if !ok {
		h.Logger.Info("Error: result not found in async map", "key", asyncKey)
		return false, errors.New("result not found in async map")
	}

	res, ok := resVal.(*asyncResult[*wrpc.Result[document.MutationMetadata, document.DocumentError]])
	if !ok {
		h.Logger.Info("Error: failed to cast map value to asyncResult", "result", res)
		return false, errors.New("failed to case map value to asyncResult")
	}
	return res.ready, nil
}

func (h *Handler) InsertResultAsync_Get(ctx__ context.Context, self wrpc.Borrow[document.InsertResultAsync]) (*wrpc.Result[document.MutationMetadata, document.DocumentError], error) {
	h.Logger.Info("InsertResultAsync Get", "self", string(self))

	asyncKey := string(self)
	resVal, ok := h.asyncResult[asyncKey]
	if !ok {
		h.Logger.Info("Error: result not found in async map", "key", asyncKey)
		return nil, errors.New("result not found in async map")
	}

	res, ok := resVal.(*asyncResult[*wrpc.Result[document.MutationMetadata, document.DocumentError]])
	if !ok {
		h.Logger.Info("Error: failed to cast map value to asyncResult", "result", res)
		return nil, errors.New("failed to case map value to asyncResult")
	}
	return res.result, nil
}

func (h *Handler) GetAsync(ctx context.Context, id string, options *document.DocumentGetOptions) (wrpc.Own[document.GetResultAsync], error) {
	h.Logger.Info("Handling GetAsync")
	res := &asyncResult[*wrpc.Result[document.DocumentGetResult, document.DocumentError]]{ready: false}
	asyncKey := fmt.Sprintf("get.%s.%s", id, uuid.NewString())
	h.asyncResult[asyncKey] = res

	res.HandleAsync(func() *wrpc.Result[document.DocumentGetResult, document.DocumentError] {
		res, err := h.Get(ctx, id, options)
		if err != nil {
			h.Logger.Error("Error InsertAsync failed", "error", err)
			return Err[document.DocumentGetResult](*types.NewDocumentErrorNotFound())
		}
		return res
	})
	return wrpc.Own[document.GetResultAsync](asyncKey), nil
}

func (h *Handler) GetResultAsync_Ready(ctx__ context.Context, self wrpc.Borrow[document.GetResultAsync]) (bool, error) {
	h.Logger.Info("GetResultAsync Ready", "self", self)
	asyncKey := string(self)

	resVal, ok := h.asyncResult[asyncKey]
	if !ok {
		h.Logger.Info("Error: result not found in async map", "key", asyncKey)
		return false, errors.New("result not found in async map")
	}

	res, ok := resVal.(*asyncResult[*wrpc.Result[document.DocumentGetResult, document.DocumentError]])
	if !ok {
		h.Logger.Info("Error: failed to cast map value to asyncResult", "result", res)
		return false, errors.New("failed to case map value to asyncResult")
	}
	return res.ready, nil
}

func (h *Handler) GetResultAsync_Get(ctx__ context.Context, self wrpc.Borrow[document.GetResultAsync]) (*wrpc.Result[document.DocumentGetResult, document.DocumentError], error) {
	h.Logger.Info("GetResultAsync Get", "self", self)
	asyncKey := string(self)

	resVal, ok := h.asyncResult[asyncKey]
	if !ok {
		h.Logger.Info("Error: result not found in async map", "key", asyncKey)
		return nil, errors.New("result not found in async map")
	}

	res, ok := resVal.(*asyncResult[*wrpc.Result[document.DocumentGetResult, document.DocumentError]])
	if !ok {
		h.Logger.Info("Error: failed to cast map value to asyncResult", "result", res)
		return nil, errors.New("failed to case map value to asyncResult")
	}
	return res.result, nil
}
