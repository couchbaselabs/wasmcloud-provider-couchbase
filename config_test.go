package main

import (
	"testing"

	"go.wasmcloud.dev/provider"
)

func TestGetConfigValue(t *testing.T) {
	config := map[string]string{
		"username":         "testuser",
		"connectionString": "couchbase://localhost",
	}
	password := &provider.SecretValue{}
	// Unfortunate way to construct a secret value
	err := password.UnmarshalJSON([]byte(`{"kind": "String", "value": "secretpassword"}`))
	if err != nil {
		t.Fatalf("failed to unmarshal password secret: %v", err)
	}
	secrets := map[string]provider.SecretValue{
		"password": *password,
	}

	tests := []struct {
		key         string
		expected    string
		expectError bool
	}{
		{"username", "testuser", false},
		{"connectionString", "couchbase://localhost", false},
		{"password", "secretpassword", false},
		{"nonexistent", "", true},
	}

	for _, test := range tests {
		value, err := getConfigValue(config, secrets, test.key)
		if test.expectError {
			if err == nil {
				t.Errorf("expected error for key '%s', got none", test.key)
			}
		} else {
			if err != nil {
				t.Errorf("did not expect error for key '%s', got %v", test.key, err)
			}
			if value != test.expected {
				t.Errorf("expected value '%s' for key '%s', got '%s'", test.expected, test.key, value)
			}
		}
	}
}
