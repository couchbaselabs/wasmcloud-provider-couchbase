package main

import (
	"errors"

	"github.com/wasmCloud/provider-sdk-go"
)

type CouchbaseConnectionArgs struct {
	Username         string
	Password         string
	BucketName       string
	Host             string
	ConnectionString string
	ScopeName        string
	CollectionName   string
}

// Construct Couchbase connection args from config and secrets
func validateCouchbaseConfig(config map[string]string, secrets map[string]provider.SecretValue) (CouchbaseConnectionArgs, error) {
	connectionArgs := CouchbaseConnectionArgs{}
	if username, ok := config["username"]; !ok || username == "" {
		return connectionArgs, errors.New("username is required")
	} else {
		connectionArgs.Username = username
	}
	if bucketName, ok := config["bucket_name"]; !ok || bucketName == "" {
		return connectionArgs, errors.New("bucket_name is required")
	} else {
		connectionArgs.BucketName = bucketName
	}
	if host, ok := config["host"]; !ok || host == "" {
		return connectionArgs, errors.New("host is required")
	} else {
		connectionArgs.Host = host
	}
	if connectionString, ok := config["connection_string"]; !ok || connectionString == "" {
		return connectionArgs, errors.New("connection_string is required")
	} else {
		connectionArgs.ConnectionString = connectionString
	}
	if scopeName, ok := config["scope_name"]; !ok || scopeName == "" {
		return connectionArgs, errors.New("scope_name is required")
	} else {
		connectionArgs.ScopeName = scopeName
	}
	if collectionName, ok := config["collection_name"]; !ok || collectionName == "" {
		return connectionArgs, errors.New("collection_name is required")
	} else {
		connectionArgs.CollectionName = collectionName
	}

	password := secrets["password"].StringValue()
	if password == "" {
		return connectionArgs, errors.New("password is required")
	} else {
		connectionArgs.Password = password
	}
	return connectionArgs, nil
}
