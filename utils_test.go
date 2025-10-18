package pocketbase

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEnvIsTruthy tests the EnvIsTruthy utility function
// This is a true unit test - no server required
func TestEnvIsTruthy(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		expected bool
	}{
		{"true string", "TEST_TRUE", "true", true},
		{"TRUE string", "TEST_TRUE_UPPER", "TRUE", true},
		{"1 string", "TEST_ONE", "1", true},
		{"yes string", "TEST_YES", "yes", true},
		{"YES string", "TEST_YES_UPPER", "YES", true},
		{"false string", "TEST_FALSE", "false", false},
		{"0 string", "TEST_ZERO", "0", false},
		{"no string", "TEST_NO", "no", false},
		{"empty string", "TEST_EMPTY", "", false},
		{"random string", "TEST_RANDOM", "random", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if err := os.Setenv(tt.envKey, tt.envValue); err != nil {
				t.Fatalf("Failed to set environment variable: %v", err)
			}
			defer func() {
				if err := os.Unsetenv(tt.envKey); err != nil {
					t.Logf("Failed to unset environment variable: %v", err)
				}
			}()

			// Test the function
			result := EnvIsTruthy(tt.envKey)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEnvIsTruthy_NonExistentKey tests behavior with non-existent environment variables
func TestEnvIsTruthy_NonExistentKey(t *testing.T) {
	result := EnvIsTruthy("NON_EXISTENT_KEY_12345")
	assert.False(t, result, "Non-existent environment variable should return false")
}

// TestGetZIPName tests the getZIPName utility function from backup.go
// This is a true unit test - no server required
func TestGetZIPName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"already has .zip", "backup.zip", "backup.zip"},
		{"already has .ZIP", "backup.ZIP", "backup.zip"},
		{"no extension", "backup", "backup.zip"},
		{"mixed case no extension", "BackUp", "backup.zip"},
		{"with other extension", "backup.tar", "backup.tar.zip"},
		{"empty string", "", ".zip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getZIPName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
