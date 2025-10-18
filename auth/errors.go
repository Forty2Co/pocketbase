package auth

import "errors"

// ErrInvalidResponse is returned when PocketBase returns an invalid response.
var ErrInvalidResponse = errors.New("invalid response")