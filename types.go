// Package pocketbase provides common types and structures used across the PocketBase Go SDK.
package pocketbase

import "github.com/Forty2Co/pocketbase/collections"

// Common response and parameter types re-exported from sub-packages for convenience and backward compatibility.
// These types are the core data structures used throughout the PocketBase SDK.

type (
	// ResponseList represents a paginated list response from PocketBase.
	// Re-exported from collections package for backward compatibility.
	ResponseList[T any] = collections.ResponseList[T]
	
	// ResponseCreate represents the response from creating a new record.
	// Re-exported from collections package for backward compatibility.
	ResponseCreate = collections.ResponseCreate
	
	// ParamsList represents query parameters for PocketBase API requests including pagination, filtering, and sorting.
	// Re-exported from collections package for backward compatibility.
	ParamsList = collections.ParamsList
)