package pocketbase

// ParamsList represents query parameters for PocketBase API requests including pagination, filtering, and sorting.
type ParamsList struct {
	Page    int
	Size    int
	Filters string
	Sort    string
	Expand  string
	Fields  string

	hackResponseRef any //hack for collection list
}
