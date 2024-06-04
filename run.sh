#!/bin/bash

host_data='
{
    "lattice_rpc_url": "0.0.0.0:4222",
    "lattice_rpc_prefix": "default",
    "provider_key": "couchbase",
    "link_name": "default",
    "config": {
        "username": "Administrator",
        "password": "password",
        "bucketName": "test",
        "connectionString": "localhost"
    }
}'
echo $host_data | base64 | go run ./
