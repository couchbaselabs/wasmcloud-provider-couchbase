package provider

import (
	// Generated bindings
	"time"

	"github.com/couchbase-examples/wasmcloud-provider-couchbase/bindings/exports/wasmcloud/couchbase/document"
	"github.com/couchbase/gocb/v2"
)

// This file contains the conversion functions for the options used in the document binding.
// TODO: complete options conversion

// GetAllReplicaOptions
func GetAllReplicaOptions(o *document.DocumentGetAllReplicaOptions) *gocb.GetAllReplicaOptions {
	if o == nil {
		return nil
	}
	return &gocb.GetAllReplicaOptions{
		Timeout: time.Duration(*o.TimeoutNs),
	}
}

// GetOptions
func GetOptions(o *document.DocumentGetOptions) *gocb.GetOptions {
	if o == nil {
		return &gocb.GetOptions{Transcoder: gocb.NewRawStringTranscoder()}
	}
	return &gocb.GetOptions{
		Transcoder: gocb.NewRawStringTranscoder(),
		WithExpiry: o.WithExpiry,
		Project:    o.Project,
		// ...
	}
}

// GetAndLockOptions
func GetAndLockOptions(o *document.DocumentGetAndLockOptions) *gocb.GetAndLockOptions {
	if o == nil {
		return nil
	}
	return &gocb.GetAndLockOptions{
		Timeout: time.Duration(*o.TimeoutNs),
	}
}

// GetAndTouchOptions
func GetAndTouchOptions(o *document.DocumentGetAndTouchOptions) *gocb.GetAndTouchOptions {
	if o == nil {
		return nil
	}
	return &gocb.GetAndTouchOptions{
		Timeout: time.Duration(*o.TimeoutNs),
	}
}

// GetAnyReplicaOptions
func GetAnyReplicaOptions(o *document.DocumentGetAnyReplicaOptions) *gocb.GetAnyReplicaOptions {
	if o == nil {
		return nil
	}
	return &gocb.GetAnyReplicaOptions{
		Timeout: time.Duration(*o.TimeoutNs),
	}
}

// InsertOptions
func InsertOptions(o *document.DocumentInsertOptions) *gocb.InsertOptions {
	if o == nil {
		return &gocb.InsertOptions{Transcoder: gocb.NewRawStringTranscoder()}
	}
	return &gocb.InsertOptions{
		Timeout:    time.Duration(*o.TimeoutNs),
		Transcoder: gocb.NewRawStringTranscoder(),
	}
}

// RemoveOptions
func RemoveOptions(o *document.DocumentRemoveOptions) *gocb.RemoveOptions {
	if o == nil {
		return nil
	}
	return &gocb.RemoveOptions{
		Timeout: time.Duration(*o.TimeoutNs),
	}
}

// ReplaceOptions
func ReplaceOptions(o *document.DocumentReplaceOptions) *gocb.ReplaceOptions {
	if o == nil {
		return &gocb.ReplaceOptions{Transcoder: gocb.NewRawStringTranscoder()}
	}
	return &gocb.ReplaceOptions{
		Timeout:    time.Duration(*o.TimeoutNs),
		Transcoder: gocb.NewRawStringTranscoder(),
	}
}

// TouchOptions
func TouchOptions(o *document.DocumentTouchOptions) *gocb.TouchOptions {
	if o == nil {
		return nil
	}
	return &gocb.TouchOptions{
		Timeout: time.Duration(*o.TimeoutNs),
	}
}

// UnlockOptions
func UnlockOptions(o *document.DocumentUnlockOptions) *gocb.UnlockOptions {
	if o == nil {
		return nil
	}
	return &gocb.UnlockOptions{
		Timeout: time.Duration(*o.TimeoutNs),
	}
}

// UpsertOptions
func UpsertOptions(o *document.DocumentUpsertOptions) *gocb.UpsertOptions {
	if o == nil {
		return &gocb.UpsertOptions{Transcoder: gocb.NewRawStringTranscoder()}
	}
	return &gocb.UpsertOptions{}
}
