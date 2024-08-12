# wasmcloud-provider-couchbase

This is a capability provider for wasmCloud to provide Couchbase KV connectivity to Wasm applications via the `wasmcloud-couchbase` interface.

This provider uses the **RawJSONTranscoder** for Couchbase, storing any new keys as binary data. The deserialization into a `struct` or structured data must be done on the component side as the provider does not have information about the desired shape of data.

## Interface support

- [x] wasmcloud:couchbase/document@0.1.0-draft
- [ ] wasmcloud:couchbase/fts@0.1.0-draft
- [ ] wasmcloud:couchbase/subdocument-lookup@0.1.0-draft
- [ ] wasmcloud:couchbase/subdocument-mutate@0.1.0-draft
- [ ] wasmcloud:couchbase/sqlpp@0.1.0-draft

## Build

Prerequisites:

- [wash 0.30](https://wasmcloud.com/docs/installation) or later

Build this capability provider with:

```shell
wash build
```

Build the example components with:

```shell
wash build -p components/rust
wash build -p components/golang
```

## Run

### Prerequisites

- [wash 0.30](https://wasmcloud.com/docs/installation) or later
- The [secrets-nats-kv](https://github.com/wasmCloud/wasmCloud/tree/main/crates/secrets-nats-kv) CLI installed (for now this requires a Rust toolchain)
- A built couchbase capability provider, see [#build](#build)
- Setup Couchbase server with the required configuration for testing using docker-compose.yaml in the repo.

```bash
docker-compose up -d
```

Alternatively, you can use [Quick Install](https://docs.couchbase.com/server/current/getting-started/do-a-quick-install.html) guide with a bucket named **test** created.

### Running

```shell
WASMCLOUD_SECRETS_TOPIC=wasmcloud.secrets \
    wash up -d

# Generate encryption keys and run the backend
export ENCRYPTION_XKEY_SEED=$(wash keys gen curve -o json | jq -r '.seed')
export TRANSIT_XKEY_SEED=$(wash keys gen curve -o json | jq -r '.seed')
secrets-nats-kv run &
# Put the password in the NATS KV secrets backend
provider_key=$(wash inspect ./build/wasmcloud-provider-couchbase.par.gz -o json | jq -r '.service')
secrets-nats-kv put couchbase_password --string password
secrets-nats-kv add-mapping $provider_key --secret couchbase_password
wash app deploy ./wadm.yaml
```

Then you can invoke either the Rust component or the Golang component on port 8080 and 8081, respectively. Both components implement the exact same functionality, which simply takes a payload and a path to roundtrip a JSON document in Couchbase.

```shell
curl -d '{"demo": true, "couchbase": "db", "wasmcloud": "application platform"}' \
    localhost:8080/mykey
{"demo": true, "couchbase": "db", "wasmcloud": "application platform"}%
```

## Test

To test the WIT bindings, download [wit-bindgen](https://github.com/bytecodealliance/wit-bindgen) and run the following:

```console
wit-deps && wit-bindgen rust --out-dir /tmp/wit wit/
```

This will attempt to generate Rust based bindings, in a folder under `/tmp` (which will be cleaned up eventually), but in doing so, will check that the WIT definitions are valid (as they must be to complete binding generation).
