package main

import (
	"strings"

	gen "github.com/wasmcloud/wasmcloud/examples/golang/components/http-hello-world/gen"
)

// Helper type aliases to make code more readable
type HttpRequest = gen.ExportsWasiHttp0_2_0_IncomingHandlerIncomingRequest
type HttpResponseWriter = gen.ExportsWasiHttp0_2_0_IncomingHandlerResponseOutparam
type HttpOutgoingResponse = gen.WasiHttp0_2_0_TypesOutgoingResponse
type HttpError = gen.WasiHttp0_2_0_TypesErrorCode

type Document = gen.WasmcloudCouchbase0_1_0_draft_DocumentDocument

type HttpServer struct{}

func init() {
	httpserver := HttpServer{}
	// Set the incoming handler struct to HttpServer
	gen.SetExportsWasiHttp0_2_0_IncomingHandler(httpserver)
}

func (h HttpServer) Handle(request HttpRequest, responseWriter HttpResponseWriter) {
	// Construct HttpResponse to send back
	headers := gen.NewFields()
	httpResponse := gen.NewOutgoingResponse(headers)
	httpResponse.SetStatusCode(200)
	body := httpResponse.Body().Unwrap()
	bodyWrite := body.Write().Unwrap()

	// Get document body & id
	documentContents := request.Consume().Unwrap().Stream().Unwrap().BlockingRead(99999).Unwrap()
	var documentId string
	if request.PathWithQuery().IsSome() {
		trimmed := strings.TrimLeft(request.PathWithQuery().Unwrap(), "/")
		if trimmed == "" {
			documentId = "demodoc"
		} else {
			documentId = trimmed
		}
	} else {
		documentId = "demodoc"
	}

	// Upsert and get the document
	document := gen.WasmcloudCouchbase0_1_0_draft_TypesDocumentRaw(string(documentContents))
	gen.WasmcloudCouchbase0_1_0_draft_DocumentUpsert(documentId, document, gen.None[gen.WasmcloudCouchbase0_1_0_draft_DocumentDocumentUpsertOptions]())
	res := gen.WasmcloudCouchbase0_1_0_draft_DocumentGet(documentId, gen.None[gen.WasmcloudCouchbase0_1_0_draft_DocumentDocumentGetOptions]())

	// Send HTTP response
	okResponse := gen.Ok[HttpOutgoingResponse, HttpError](httpResponse)
	gen.StaticResponseOutparamSet(responseWriter, okResponse)
	if res.IsOk() {
		document := res.Unwrap()
		bodyWrite.BlockingWriteAndFlush([]uint8(document.Document.GetRaw())).Unwrap()
	} else {
		bodyWrite.BlockingWriteAndFlush([]uint8("Document not found")).Unwrap()
	}
	bodyWrite.Drop()
	gen.StaticOutgoingBodyFinish(body, gen.None[gen.WasiHttp0_2_0_TypesTrailers]())
}

//go:generate wit-bindgen tiny-go wit --out-dir=gen --gofmt
func main() {}
