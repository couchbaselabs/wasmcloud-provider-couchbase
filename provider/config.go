package provider

import (
	"errors"

	"go.wasmcloud.dev/provider"
)

type CouchbaseConnectionArgs struct {
	Username         string
	Password         string
	BucketName       string
	ConnectionString string
	ScopeName        string
	CollectionName   string
}

// Construct Couchbase connection args from config and secrets
func validateCouchbaseConfig(config map[string]string, secrets map[string]provider.SecretValue) (CouchbaseConnectionArgs, error) {
	connectionArgs := CouchbaseConnectionArgs{}
	if username, ok := config["username"]; !ok || username == "" {
		return connectionArgs, errors.New("username config is required")
	} else {
		connectionArgs.Username = username
	}
	if bucketName, ok := config["bucketName"]; !ok || bucketName == "" {
		return connectionArgs, errors.New("bucketName config is required")
	} else {
		connectionArgs.BucketName = bucketName
	}
	if connectionString, ok := config["connectionString"]; !ok || connectionString == "" {
		return connectionArgs, errors.New("connectionString config is required")
	} else {
		connectionArgs.ConnectionString = connectionString
	}

	password := secrets["password"].String.Reveal()
	if password == "" {
		return connectionArgs, errors.New("password secret is required")
	} else {
		connectionArgs.Password = password
	}
	return connectionArgs, nil
}
