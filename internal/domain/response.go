package domain

type APIResponse struct {
	Status           string      `json:"status"`
	StatusCode       int         `json:"statusCode"`
	TrackingID       string      `json:"trackingId"`
	Data             interface{} `json:"data,omitempty"`
	Error            *APIError   `json:"error,omitempty"`
	DocumentationURL string      `json:"documentationUrl"`
}

type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details"`
	Timestamp  string `json:"timestamp"`
	Path       string `json:"path"`
	Suggestion string `json:"suggestion"`
}

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}
