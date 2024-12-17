package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/bytecodealliance/wasm-tools-go/cm"
	"github.com/couchbaselabs/wasmcloud-provider-couchbase/components/golang/gen/wasmcloud/couchbase/document"
	"github.com/couchbaselabs/wasmcloud-provider-couchbase/components/golang/gen/wasmcloud/couchbase/types"
	"github.com/julienschmidt/httprouter"
	"go.wasmcloud.dev/component/log/wasilog"
	"go.wasmcloud.dev/component/net/wasihttp"
)

var logger = wasilog.ContextLogger("golang-hello")

func init() {
	router := httprouter.New()
	router.POST("/:document_id", handleRequest)
	wasihttp.Handle(router)
}

func handleRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get document id
	documentIdParameter := ps.ByName("document_id")
	if len(documentIdParameter) == 0 {
		documentIdParameter = "demodoc"
	}
	docId := types.DocumentID(documentIdParameter)

	// Get document body
	data, err := io.ReadAll(r.Body)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Unable to read http body",
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	doc := types.DocumentRaw(types.JSONString(data))

	// Upsert document
	upsertResult := document.Upsert(docId, doc, cm.Option[document.DocumentUpsertOptions]{})
	if upsertResult.IsErr() {
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Failed to upsert document",
			"error":   upsertResult.Err().String(),
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get document
	getResult := document.Get(docId, cm.Option[document.DocumentGetOptions]{})
	if getResult.IsErr() {
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Failed to get document",
			"error":   getResult.Err().String(),
		})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Write HTTP response
	w.Write([]byte(*getResult.OK().Document.Raw()))
	w.WriteHeader(http.StatusOK)
}

//go:generate go run github.com/bytecodealliance/wasm-tools-go/cmd/wit-bindgen-go generate --world hello --out gen ./wit
func main() {}
