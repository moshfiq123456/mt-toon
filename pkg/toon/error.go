package toon

import "fmt"

// ErrCode represents standardized error codes
type ErrCode string

const (
	ErrCodeInvalidResponse    ErrCode = "INVALID_RESPONSE"
	ErrCodeEmptyResponse      ErrCode = "EMPTY_RESPONSE"
	ErrCodeJSONUnmarshal      ErrCode = "JSON_UNMARSHAL"
	ErrCodeNilHandler         ErrCode = "NIL_HANDLER"
	ErrCodeNilResponse        ErrCode = "NIL_RESPONSE"
	ErrCodeEmptyData          ErrCode = "EMPTY_DATA"
	ErrCodeIORead             ErrCode = "IO_READ"
	ErrCodeInvalidStatusCode  ErrCode = "INVALID_STATUS_CODE"
)

// ValidationError represents a validation error with context
type ValidationError struct {
	Code    ErrCode
	Message string
	Err     error
	Context map[string]interface{}
}

// Error implements the error interface for ValidationError
func (ve *ValidationError) Error() string {
	if ve == nil {
		return ""
	}
	if ve.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", ve.Code, ve.Message, ve.Err)
	}
	return fmt.Sprintf("[%s] %s", ve.Code, ve.Message)
}

// Unwrap returns the underlying error
func (ve *ValidationError) Unwrap() error {
	if ve == nil {
		return nil
	}
	return ve.Err
}
