//go:generate go run github.com/bytecodealliance/wasm-tools-go/cmd/wit-bindgen-go generate --world hello --out gen ./wit

package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bytecodealliance/wasm-tools-go/cm"
	"github.com/wasmcloud/wasmcloud/examples/golang/components/http-hello-world/gen/wasmcloud/couchbase/v0.1.0-draft/document"
	"github.com/wasmcloud/wasmcloud/examples/golang/components/http-hello-world/gen/wasmcloud/couchbase/v0.1.0-draft/types"
	"go.wasmcloud.dev/component/net/wasihttp"
)

func init() {
	wasihttp.HandleFunc(documentHandler)
}

func documentHandler(w http.ResponseWriter, r *http.Request) {
	// Get document body & id
	documentId := "demodoc"
	trimmed := strings.TrimLeft(r.URL.Path, "/")
	if trimmed != "" {
		documentId = trimmed
	}

	var input map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Provided input could not be decoded, please check that your input is valid JSON.", http.StatusBadRequest)
		return
	}

	output, err := json.Marshal(input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Upsert and get the document
	doc := types.DocumentRaw(types.JSONString(string(output)))
	upsert := document.Upsert(types.DocumentID(documentId), doc, cm.None[document.DocumentUpsertOptions]())
	if upsert.IsErr() {
		http.Error(w, upsert.Err().String(), http.StatusInternalServerError)
		return
	}

	res := document.Get(types.DocumentID(documentId), cm.None[document.DocumentGetOptions]())
	if res.IsErr() {
		http.Error(w, res.Err().String(), http.StatusInternalServerError)
		return
	}

	// Send HTTP response
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(*res.OK().Document.Raw())
}

func main() {}
