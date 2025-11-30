package toon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Handler processes Toon API responses and provides convenient methods for access
// Handler is safe for concurrent use after initialization
type Handler struct {
	resp   *Response
	body   []byte
	rawErr error
	mu     sync.RWMutex
}

// NewHandler creates a new Handler from raw bytes
// It performs comprehensive validation and error handling
func NewHandler(body []byte) (*Handler, error) {
	if body == nil {
		return nil, &ValidationError{
			Code:    ErrCodeEmptyResponse,
			Message: "body is nil",
		}
	}

	if len(body) == 0 {
		return nil, &ValidationError{
			Code:    ErrCodeEmptyResponse,
			Message: "body is empty",
		}
	}

	var resp Response
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, &ValidationError{
			Code:    ErrCodeJSONUnmarshal,
			Message: "failed to unmarshal response body",
			Err:     err,
			Context: map[string]interface{}{
				"body_size": len(body),
			},
		}
	}

	return &Handler{
		resp: &resp,
		body: body,
	}, nil
}

// FromHTTPResponse creates a Handler from an HTTP response
// It validates the response, reads the body, and handles errors comprehensively
func FromHTTPResponse(httpResp *http.Response) (*Handler, error) {
	if httpResp == nil {
		return nil, &ValidationError{
			Code:    ErrCodeInvalidResponse,
			Message: "http response is nil",
		}
	}

	// Ensure body is closed
	if httpResp.Body != nil {
		defer func() {
			_ = httpResp.Body.Close()
		}()
	}

	if httpResp.Body == nil {
		return nil, &ValidationError{
			Code:    ErrCodeInvalidResponse,
			Message: "http response body is nil",
			Context: map[string]interface{}{
				"status_code": httpResp.StatusCode,
			},
		}
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, &ValidationError{
			Code:    ErrCodeIORead,
			Message: "failed to read response body",
			Err:     err,
			Context: map[string]interface{}{
				"status_code": httpResp.StatusCode,
			},
		}
	}

	handler, err := NewHandler(body)
	if err != nil {
		return nil, err
	}

	// Validate HTTP status code against response success flag
	if (httpResp.StatusCode < 200 || httpResp.StatusCode >= 300) && handler.IsSuccess() {
		return nil, &ValidationError{
			Code:    ErrCodeInvalidStatusCode,
			Message: "http status code indicates error but response success is true",
			Context: map[string]interface{}{
				"status_code": httpResp.StatusCode,
				"success":     handler.IsSuccess(),
			},
		}
	}

	return handler, nil
}

// IsSuccess safely checks if the response indicates success
func (h *Handler) IsSuccess() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h == nil || h.resp == nil {
		return false
	}
	return h.resp.Success
}

// IsError safely checks if the response contains an error
func (h *Handler) IsError() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h == nil || h.resp == nil {
		return true
	}
	return !h.resp.Success && h.resp.Error != nil
}

// GetError safely returns the error from the response, if present
func (h *Handler) GetError() *ResponseError {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h == nil || h.resp == nil {
		return nil
	}
	return h.resp.Error
}

// ErrorString returns a formatted error string
// Returns empty string if no error is present
func (h *Handler) ErrorString() string {
	err := h.GetError()
	if err == nil {
		return ""
	}

	parts := []string{err.Code}
	if err.Message != "" {
		parts = append(parts, err.Message)
	}
	if err.Details != "" {
		parts = append(parts, err.Details)
	}
	if err.Field != "" {
		parts = append(parts, fmt.Sprintf("field: %s", err.Field))
	}

	result := ""
	for i, part := range parts {
		if i == 0 {
			result = part
		} else {
			result += " | " + part
		}
	}
	return result
}

// GetData safely returns the raw data from the response
func (h *Handler) GetData() json.RawMessage {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h == nil || h.resp == nil {
		return nil
	}

	if len(h.resp.Data) == 0 {
		return nil
	}

	// Return a copy to prevent external modification
	data := make(json.RawMessage, len(h.resp.Data))
	copy(data, h.resp.Data)
	return data
}

// UnmarshalData safely unmarshals the response data into the provided interface
// Returns ValidationError if data is empty or unmarshal fails
func (h *Handler) UnmarshalData(v interface{}) error {
	if v == nil {
		return &ValidationError{
			Code:    ErrCodeInvalidResponse,
			Message: "target interface is nil",
		}
	}

	data := h.GetData()
	if len(data) == 0 {
		return &ValidationError{
			Code:    ErrCodeEmptyData,
			Message: "response data is empty",
		}
	}

	if err := json.Unmarshal(data, v); err != nil {
		return &ValidationError{
			Code:    ErrCodeJSONUnmarshal,
			Message: "failed to unmarshal data into target type",
			Err:     err,
			Context: map[string]interface{}{
				"data_size": len(data),
				"target":    fmt.Sprintf("%T", v),
			},
		}
	}

	return nil
}

// GetMeta safely returns the metadata from the response
func (h *Handler) GetMeta() *Meta {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h == nil || h.resp == nil {
		return nil
	}
	return h.resp.Meta
}

// GetRequestID safely returns the request ID from metadata if available
func (h *Handler) GetRequestID() string {
	meta := h.GetMeta()
	if meta == nil {
		return ""
	}
	return meta.RequestID
}

// GetRateLimit safely returns rate limit information if available
func (h *Handler) GetRateLimit() *RateLimit {
	meta := h.GetMeta()
	if meta == nil {
		return nil
	}
	return meta.RateLimit
}

// IsRateLimited checks if the request was rate limited based on remaining quota
func (h *Handler) IsRateLimited() bool {
	rl := h.GetRateLimit()
	if rl == nil {
		return false
	}
	return rl.Remaining <= 0
}

// GetRateLimitReset safely returns the rate limit reset time
func (h *Handler) GetRateLimitReset() *time.Time {
	rl := h.GetRateLimit()
	if rl == nil {
		return nil
	}
	return &rl.Reset
}

// GetRateLimitStatus returns a formatted status of rate limits
func (h *Handler) GetRateLimitStatus() string {
	rl := h.GetRateLimit()
	if rl == nil {
		return "rate limit information not available"
	}

	remaining := rl.Remaining
	if remaining < 0 {
		remaining = 0
	}

	return fmt.Sprintf("%d/%d requests remaining (reset: %s)",
		remaining, rl.Limit, rl.Reset.Format(time.RFC3339))
}

// GetAPIVersion safely returns the API version from metadata
func (h *Handler) GetAPIVersion() string {
	meta := h.GetMeta()
	if meta == nil {
		return ""
	}
	return meta.APIVersion
}

// GetTimestamp safely returns the response timestamp
func (h *Handler) GetTimestamp() *time.Time {
	meta := h.GetMeta()
	if meta == nil || meta.Timestamp.IsZero() {
		return nil
	}
	return &meta.Timestamp
}

// String returns a formatted string representation of the response
func (h *Handler) String() string {
	if h == nil || h.resp == nil {
		return "Handler(nil)"
	}

	if h.resp.Success {
		requestID := h.GetRequestID()
		if requestID != "" {
			return fmt.Sprintf("Handler(Success, RequestID=%s)", requestID)
		}
		return "Handler(Success)"
	}

	errStr := h.ErrorString()
	if errStr != "" {
		return fmt.Sprintf("Handler(Error=%s)", errStr)
	}
	return "Handler(Error)"
}

// RawBody returns the original unparsed response body
func (h *Handler) RawBody() []byte {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h == nil || h.body == nil {
		return nil
	}

	// Return a copy to prevent external modification
	body := make([]byte, len(h.body))
	copy(body, h.body)
	return body
}

// Response returns the underlying Response struct
// Callers should not modify the returned struct
func (h *Handler) Response() *Response {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h == nil {
		return nil
	}
	return h.resp
}
