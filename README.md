# wasmcloud-provider-couchbase

This is a capability provider for wasmCloud to provide Couchbase KV connectivity to Wasm applications via `wasi-keyvalue`. At the moment it supports the `wasi:keyvalue/store@0.2.0-draft` interface.

This provider uses the **RawJSONTranscoder** for Couchbase, storing any new keys as binary data. Since the wasi-keyvalue interface works entirely in storing and retrieving binary data, the deserialization into a `struct` or structured data must be done on the component side.

## Build

Prerequisites:

- [wash 0.29](https://wasmcloud.com/docs/installation) or later

Build this capability provider with:

```shell
wash build
```

## Run

Prerequisites:

- [wash 0.29](https://wasmcloud.com/docs/installation) or later
- A built couchbase capability provider, see [#build](#build)
- Couchbase server as setup in the [Quick Install](https://docs.couchbase.com/server/current/getting-started/do-a-quick-install.html) guide with a bucket named **test** created.

```shell
wash up -d
wash app deploy ./wadm.yaml
```

Then you can test the increment functionality with cURL:

```shell
curl localhost:8080/couchbase
```

## Test

To test the WIT bindings, download [`wit-bindgen`][wit-bindgen] and run the following:

```console
wit-deps && wit-bindgen rust --out-dir /tmp/wit wit/
```

This will attempt to generate Rust based bindings, in a folder under `/tmp` (which will be cleaned up eventually), but in doing so, will check that the WIT definitions are valid (as they must be to complete binding generation).
