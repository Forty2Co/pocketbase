package pocketbase

// ResponseList represents a paginated list response from PocketBase.
type ResponseList[T any] struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
	Items      []T `json:"items"`
}

// ResponseCreate represents the response from creating a new record.
type ResponseCreate struct {
	ID      string `json:"id"`
	Created string `json:"created"`
	Field   string `json:"field"`
	Updated string `json:"updated"`
}
