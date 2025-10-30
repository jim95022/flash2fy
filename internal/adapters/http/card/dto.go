package cardhttp

// cardRequest transports card creation/update payloads from HTTP.
type cardRequest struct {
	Front string `json:"front"`
	Back  string `json:"back"`
}

// cardResponse captures the serialized flashcard representation returned to clients.
type cardResponse struct {
	ID        string `json:"id"`
	Front     string `json:"front"`
	Back      string `json:"back"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
