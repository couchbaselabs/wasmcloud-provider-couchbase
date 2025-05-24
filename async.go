package main

import (
	"context"
	"errors"
	"sync"

	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/exports/wasmcloud/couchbase/document"
	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/wasmcloud/couchbase/types"
	"github.com/google/uuid"
	wrpc "wrpc.io/go"
)

type asyncMap struct {
	m sync.Map
}

func HandleAsyncResult[T any](m *asyncMap, f func() T) (string, error) {
	// Create new key (this will eventually be cast to the resource key)
	asyncKey := uuid.NewString()

	// Store new asyncResult in the asyncMap using the asyncKey
	res := asyncResult[T]{ready: false}
	m.m.Store(asyncKey, &res)

	// Start
	go func() {
		res.result = f()
		res.ready = true
	}()

	// Return the asyncKey
	return asyncKey, nil
}

func getResult[T any](m *asyncMap, key string) (*asyncResult[T], error) {
	resultVal, ok := m.m.Load(key)
	if !ok {
		return nil, errors.New("result not found")
	}
	result, ok := resultVal.(*asyncResult[T])
	if !ok {
		return nil, errors.New("invalid result type")
	}
	return result, nil
}

func IsReady[T any](m *asyncMap, key string) (bool, error) {
	result, err := getResult[T](m, key)
	if err != nil {
		return false, err
	}
	return result.ready, nil
}

func Result[T any](m *asyncMap, key string) (T, error) {
	result, err := getResult[T](m, key)
	if err != nil {
		var res T
		return res, err
	}
	return result.result, err
}

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
	asyncKey, err := HandleAsyncResult(&h.asyncMap, func() *wrpc.Result[document.MutationMetadata, document.DocumentError] {
		res, err := h.Insert(ctx, id, doc, options)
		if err != nil {
			h.Logger.Error("Error InsertAsync failed", "error", err)
			return Err[types.MutationMetadata](*types.NewDocumentErrorNotFound())
		}
		return res
	})
	if err != nil {
		return nil, err
	}
	return wrpc.Own[document.InsertResultAsync](asyncKey), nil
}

func (h *Handler) InsertResultAsync_Ready(ctx__ context.Context, self wrpc.Borrow[document.InsertResultAsync]) (bool, error) {
	return IsReady[*wrpc.Result[document.MutationMetadata, document.DocumentError]](&h.asyncMap, string(self))
}

func (h *Handler) InsertResultAsync_Get(ctx__ context.Context, self wrpc.Borrow[document.InsertResultAsync]) (*wrpc.Result[document.MutationMetadata, document.DocumentError], error) {
	return Result[*wrpc.Result[document.MutationMetadata, document.DocumentError]](&h.asyncMap, string(self))
}

func (h *Handler) GetAsync(ctx context.Context, id string, options *document.DocumentGetOptions) (wrpc.Own[document.GetResultAsync], error) {
	asyncKey, err := HandleAsyncResult(&h.asyncMap, func() *wrpc.Result[document.DocumentGetResult, document.DocumentError] {
		res, err := h.Get(ctx, id, options)
		if err != nil {
			h.Logger.Error("Error InsertAsync failed", "error", err)
			return Err[document.DocumentGetResult](*types.NewDocumentErrorNotFound())
		}
		return res
	})
	if err != nil {
		return nil, err
	}
	return wrpc.Own[document.GetResultAsync](asyncKey), nil
}

func (h *Handler) GetResultAsync_Ready(ctx__ context.Context, self wrpc.Borrow[document.GetResultAsync]) (bool, error) {
	return IsReady[*wrpc.Result[document.DocumentGetResult, document.DocumentError]](&h.asyncMap, string(self))
}

func (h *Handler) GetResultAsync_Get(ctx__ context.Context, self wrpc.Borrow[document.GetResultAsync]) (*wrpc.Result[document.DocumentGetResult, document.DocumentError], error) {
	return Result[*wrpc.Result[document.DocumentGetResult, document.DocumentError]](&h.asyncMap, string(self))
}

func (h *Handler) UpsertAsync(ctx context.Context, id string, doc *types.Document, options *document.DocumentUpsertOptions) (wrpc.Own[document.UpsertResultAsync], error) {
	asyncKey, err := HandleAsyncResult(&h.asyncMap, func() *wrpc.Result[document.MutationMetadata, document.DocumentError] {
		res, err := h.Upsert(ctx, id, doc, options)
		if err != nil {
			h.Logger.Error("Error InsertAsync failed", "error", err)
			return Err[types.MutationMetadata](*types.NewDocumentErrorNotFound())
		}
		return res
	})
	if err != nil {
		return nil, err
	}
	return wrpc.Own[document.UpsertResultAsync](asyncKey), nil
}

func (h *Handler) UpsertResultAsync_Ready(ctx__ context.Context, self wrpc.Borrow[document.UpsertResultAsync]) (bool, error) {
	return IsReady[*wrpc.Result[document.MutationMetadata, document.DocumentError]](&h.asyncMap, string(self))
}

func (h *Handler) UpsertResultAsync_Get(ctx__ context.Context, self wrpc.Borrow[document.UpsertResultAsync]) (*wrpc.Result[document.MutationMetadata, document.DocumentError], error) {
	return Result[*wrpc.Result[document.MutationMetadata, document.DocumentError]](&h.asyncMap, string(self))
}
