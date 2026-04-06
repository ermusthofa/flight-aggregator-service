package domain

type APIResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}
