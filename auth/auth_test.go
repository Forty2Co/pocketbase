package auth

import (
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

const (
	defaultURL = "http://127.0.0.1:8090"
)

func TestEmailPasswordAuth_NewEmailPasswordAuth(t *testing.T) {
	client := resty.New()
	url := "http://example.com/auth"
	email := "test@example.com"
	password := "password123"

	auth := NewEmailPasswordAuth(client, url, email, password)

	assert.NotNil(t, auth)
	assert.Implements(t, (*Store)(nil), auth)
	assert.False(t, auth.IsValid()) // Should be invalid initially
	assert.Empty(t, auth.Token())   // Should have no token initially
}

func TestEmailPasswordAuth_IsValid(t *testing.T) {
	client := resty.New()
	auth := &EmailPasswordAuth{
		client:     client,
		email:      "test@example.com",
		password:   "password123",
		url:        "http://example.com/auth",
		tokenValid: time.Now().Add(30 * time.Minute), // Valid for 30 minutes
	}

	assert.True(t, auth.IsValid())

	// Test expired token
	auth.tokenValid = time.Now().Add(-30 * time.Minute) // Expired 30 minutes ago
	assert.False(t, auth.IsValid())
}

func TestEmailPasswordAuth_Token(t *testing.T) {
	client := resty.New()
	auth := &EmailPasswordAuth{
		client:   client,
		email:    "test@example.com",
		password: "password123",
		url:      "http://example.com/auth",
		token:    "test_token_123",
	}

	assert.Equal(t, "test_token_123", auth.Token())

	// Test empty token
	auth.token = ""
	assert.Empty(t, auth.Token())
}

func TestTokenAuth_NewTokenAuth(t *testing.T) {
	client := resty.New()
	url := "http://example.com/refresh"
	token := "test_token_123"

	auth := NewTokenAuth(client, url, token)

	assert.NotNil(t, auth)
	assert.Implements(t, (*Store)(nil), auth)
	assert.Equal(t, token, auth.Token())
}

func TestTokenAuth_IsValid(t *testing.T) {
	client := resty.New()
	auth := &TokenAuth{
		client:     client,
		url:        "http://example.com/refresh",
		token:      "test_token_123",
		tokenValid: time.Now().Add(30 * time.Minute), // Valid for 30 minutes
	}

	assert.True(t, auth.IsValid())

	// Test expired token
	auth.tokenValid = time.Now().Add(-30 * time.Minute) // Expired 30 minutes ago
	assert.False(t, auth.IsValid())
}

func TestTokenAuth_Token(t *testing.T) {
	client := resty.New()
	auth := &TokenAuth{
		client: client,
		url:    "http://example.com/refresh",
		token:  "test_token_456",
	}

	assert.Equal(t, "test_token_456", auth.Token())

	// Test empty token
	auth.token = ""
	assert.Empty(t, auth.Token())
}

func TestNoOpAuth(t *testing.T) {
	auth := NoOpAuth{}

	assert.NoError(t, auth.Authorize())
	assert.False(t, auth.IsValid())
	assert.Empty(t, auth.Token())
}

func TestStore_Interface(t *testing.T) {
	// Test that our implementations satisfy the Store interface
	var _ Store = &EmailPasswordAuth{}
	var _ Store = &TokenAuth{}
	var _ Store = NoOpAuth{}

	// Test that Store implementations also implement Authorizer
	var _ Authorizer = &EmailPasswordAuth{}
	var _ Authorizer = &TokenAuth{}
	var _ Authorizer = NoOpAuth{}
}