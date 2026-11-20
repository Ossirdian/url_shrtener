package handlers

type Response struct {
	Status string `json:"status"`
	Alias  string `json:"alias,omitempty"`
	URL    string `json:"url,omitempty"`
	Error  string `json:"error,omitempty"`
}

type OKResponse struct {
	Status string `json:"status"`
	Alias  string `json:"alias"`
	URL    string `json:"url"`
}

type ErrorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}
