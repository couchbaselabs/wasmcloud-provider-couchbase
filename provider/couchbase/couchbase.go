package couchbase

import (
	"log/slog"

	"github.com/couchbase-examples/wasmcloud-provider-couchbase/provider"
)

type CouchbaseHandler struct {
	provider provider.Handler
	logger   *slog.Logger
}

// couchbase.New creates a new CouchbaseHandler from a provider.Handler
func New(p *provider.Handler) *CouchbaseHandler {
	return &CouchbaseHandler{
		provider: *p,
		logger:   p.Logger,
	}
}
