package opnsense

// SearchResult is the generic paginated response from OPNsense search endpoints
type SearchResult struct {
	Rows     []map[string]any `json:"rows"`
	RowCount int              `json:"rowCount"`
	Total    int              `json:"total"`
	Current  int              `json:"current"`
}

// UUIDResult is the response from add/set operations
type UUIDResult struct {
	Result string `json:"result"`
	UUID   string `json:"uuid,omitempty"`
}

// StatusResult is a generic status response
type StatusResult struct {
	Status string `json:"status"`
}
