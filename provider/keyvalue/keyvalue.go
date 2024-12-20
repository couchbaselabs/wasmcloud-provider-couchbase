package keyvalue

import (
	"log/slog"

	"github.com/couchbase-examples/wasmcloud-provider-couchbase/provider"
)

type KeyvalueHandler struct {
	provider provider.Handler
	logger   *slog.Logger
}

// keyvalue.New creates a new KeyvalueHandler from a provider.Handler
func New(p *provider.Handler) *KeyvalueHandler {
	return &KeyvalueHandler{
		provider: *p,
		logger:   p.Logger,
	}
}
