package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetZIPName tests the GetZIPName utility function
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
			result := GetZIPName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}