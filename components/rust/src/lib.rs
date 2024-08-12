wit_bindgen::generate!({ generate_all });

use exports::wasi::http::incoming_handler::Guest;
use wasi::http::types::*;
use wasmcloud::couchbase::document as rustcb;

const KEY: &str = "demodoc";

struct HttpServer;

impl Guest for HttpServer {
    fn handle(request: IncomingRequest, response_out: ResponseOutparam) {
        let response = OutgoingResponse::new(Fields::new());
        response.set_status_code(200).unwrap();

        let body = request
            .consume()
            .unwrap()
            .stream()
            .unwrap()
            .blocking_read(u64::MAX)
            .unwrap();

        let document_id = request
            .path_with_query()
            .map(|s| {
                let id = s.trim_start_matches('/').to_string();
                if id.is_empty() {
                    KEY.to_string()
                } else {
                    id
                }
            })
            .unwrap_or_else(|| KEY.to_string());

        let document = rustcb::Document::Raw(String::from_utf8_lossy(&body).to_string());
        // Store document
        if rustcb::upsert(&document_id, document, None).is_err() {
            send_response(response, "Failed to upsert document", response_out);
            return;
        }
        // Retrieve document
        let document = rustcb::get(&document_id, None);
        let Ok(rustcb::Document::Raw(value)) = document.map(|doc| doc.document) else {
            send_response(response, "Error decoding value", response_out);
            return;
        };

        send_response(response, value, response_out);
    }
}

fn send_response(
    response: OutgoingResponse,
    bytes: impl AsRef<[u8]>,
    response_out: ResponseOutparam,
) {
    let response_body = response.body().unwrap();
    ResponseOutparam::set(response_out, Ok(response));
    response_body
        .write()
        .unwrap()
        .blocking_write_and_flush(bytes.as_ref())
        .unwrap();

    OutgoingBody::finish(response_body, None).expect("failed to finish response body");
}

export!(HttpServer);
