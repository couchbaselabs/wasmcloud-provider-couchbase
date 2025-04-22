package main

import (
	"errors"
	"fmt"

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
	username, err := getConfigValue(config, secrets, "username")
	if err != nil {
		return connectionArgs, errors.New("username config is required")
	}
	connectionArgs.Username = username

	password, err := getConfigValue(config, secrets, "password")
	if err != nil {
		return connectionArgs, errors.New("password secret is required")
	}
	connectionArgs.Password = password

	bucketName, err := getConfigValue(config, secrets, "bucketName")
	if err != nil {
		return connectionArgs, errors.New("bucketName config is required")
	}
	connectionArgs.BucketName = bucketName

	connectionString, err := getConfigValue(config, secrets, "connectionString")
	if err != nil {
		return connectionArgs, errors.New("connectionString config is required")
	}
	connectionArgs.ConnectionString = connectionString

	// scopeName and collectionName are optional
	if scopeName, err := getConfigValue(config, secrets, "scopeName"); err == nil {
		connectionArgs.ScopeName = scopeName
	}
	if collectionName, err := getConfigValue(config, secrets, "collectionName"); err == nil {
		connectionArgs.CollectionName = collectionName
	}

	return connectionArgs, nil
}

// getConfigValue retrieves the value for a given key from either the secrets map or the config map.
// It first checks the secrets map and returns the revealed secret if it's not empty.
// If not found, it checks the config map and returns the value if it's not empty.
// If the key is missing or both values are empty, it returns an error.
//
// Parameters:
//   - config: A map of configuration key-value pairs.
//   - secrets: A map of secret key-value pairs, where values are of type provider.SecretValue.
//   - key: The key to look up.
//
// Returns:
//   - The value from the secrets or config map.
//   - An error if the key is missing or both values are empty.
func getConfigValue(config map[string]string, secrets map[string]provider.SecretValue, key string) (string, error) {
	if secret, ok := secrets[key]; ok && secret.String.Reveal() != "" {
		return secret.String.Reveal(), nil
	}
	if value, ok := config[key]; ok && value != "" {
		return value, nil
	}
	return "", fmt.Errorf("key '%s' not found in config or secrets", key)
}
