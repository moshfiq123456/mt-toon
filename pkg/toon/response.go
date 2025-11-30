package toon

import (
	"encoding/json"
	"time"
)

// Response represents a standard Toon API response wrapper
type Response struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
	Meta    *Meta           `json:"meta,omitempty"`
}

// ResponseError represents error information in a Toon response
type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Field   string `json:"field,omitempty"`
}

// Meta contains metadata about the response
type Meta struct {
	Timestamp   time.Time  `json:"timestamp,omitempty"`
	RequestID   string     `json:"request_id,omitempty"`
	APIVersion  string     `json:"api_version,omitempty"`
	RateLimit   *RateLimit `json:"rate_limit,omitempty"`
}

// RateLimit contains rate limiting information
type RateLimit struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
}