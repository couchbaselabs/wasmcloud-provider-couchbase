//go:generate go tool wit-bindgen-go generate --world hello --out gen ./wit

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bytecodealliance/wasm-tools-go/cm"
	"github.com/couchbaselabs/wasmcloud-provider-couchbase/components/golang/gen/wasmcloud/couchbase/document"
	"github.com/couchbaselabs/wasmcloud-provider-couchbase/components/golang/gen/wasmcloud/couchbase/types"
	"github.com/julienschmidt/httprouter"
	"go.wasmcloud.dev/component/log/wasilog"
	"go.wasmcloud.dev/component/net/wasihttp"
)

var logger = wasilog.DefaultLogger

func init() {
	router := httprouter.New()
	router.POST("/:document_id", handleRequest)
	wasihttp.Handle(router)
}

type timeResult struct {
	Start   int64 `json:"start"`
	End     int64 `json:"end"`
	Elapsed int64 `json:"elapsed"`
}

func handleRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get document id
	documentIdParameter := ps.ByName("document_id")
	if len(documentIdParameter) == 0 {
		documentIdParameter = "demodoc"
	}

	res := []timeResult{}

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			start := time.Now()
			err := CouchbaseSomeStuff(documentIdParameter, i)
			if err != nil {
				logger.Error("error", "error", err)
			}
			end := time.Now()
			res = append(res, timeResult{
				Start:   start.UnixMicro(),
				End:     end.UnixMicro(),
				Elapsed: end.Sub(start).Microseconds(),
			})
		}()
	}
	wg.Wait()

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		logger.Error("failed to encode response body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
	}
}

func CouchbaseSomeStuff(id string, index int) error {
	documentId := types.DocumentID(fmt.Sprintf("%s.%d", id, index))

	// Insert document
	insertResult := document.UpsertAsync(documentId, types.DocumentRaw(types.JSONString("Hello World")), cm.None[document.DocumentUpsertOptions]())
	_, Err, isErr := Await(insertResult).Result()
	if isErr {
		return errors.New(Err.String())
	}

	// Get document
	getResult := document.GetAsync(documentId, cm.None[document.DocumentGetOptions]())
	getDoc, Err, isErr := Await(getResult).Result()
	if isErr {
		return errors.New(Err.String())
	}
	logger.Info("CouchbaseSomeStuff", "document", *getDoc.Document.Raw(), "index", index)
	return nil
}

//go:generate go run github.com/bytecodealliance/wasm-tools-go/cmd/wit-bindgen-go generate --world hello --out gen ./wit
func main() {}
