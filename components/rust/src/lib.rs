wit_bindgen::generate!({ generate_all });

use std::io::Read;

use wasmcloud_component::http::{self, ErrorCode};
use wasmcloud::couchbase::document as rustcb;

const KEY: &str = "demodoc";

struct Component;

http::export!(Component);

impl http::Server for Component {
    fn handle(
        request: http::IncomingRequest
    ) -> http::Result<http::Response<impl http::OutgoingBody>> {
        let (path, mut body) = request.into_parts();

        // Get document id
        let document_id = path.uri
            .path_and_query()
            .map(|s| {
                let id = s.path().trim_start_matches('/').to_string();
                if id.is_empty() {
                    KEY.to_string()
                } else {
                    id
                }
            })
            .unwrap_or_else(|| KEY.to_string());

        // Get document body
        let mut document_body = Vec::new();
        if body.read_to_end(&mut document_body).is_err() {
            return http::Response::builder()
                .status(500)
                .body("Failed to read http body".into())
                .map_err(|e| {
                    ErrorCode::InternalError(Some(format!("failed to build response {e:?}")))
                });
        }
        let document = rustcb::Document::Raw(String::from_utf8(document_body).expect("Failed to read http body"));

        // Upsert document
        if rustcb::upsert(&document_id, document, None).is_err() {
            return http::Response::builder()
                .status(500)
                .body("Failed to upsert document".into())
                .map_err(|e| {
                    ErrorCode::InternalError(Some(format!("failed to build response {e:?}")))
                });
        }

        // Get document
        let res = rustcb::get(&document_id, None);
        if res.is_err() {
            return http::Response::builder()
                .status(500)
                .body("Failed to get document".into())
                .map_err(|e| {
                    ErrorCode::InternalError(Some(format!("failed to build response {e:?}")))
                });
        }
        let Ok(rustcb::Document::Raw(value)) = res.map(|doc| doc.document) else {
            return http::Response::builder()
                .status(500)
                .body("Failed to read get document".into())
                .map_err(|e| {
                    ErrorCode::InternalError(Some(format!("failed to build response {e:?}")))
                });
        };

        // Write HTTP response
        Ok(http::Response::new(value))
    }
}