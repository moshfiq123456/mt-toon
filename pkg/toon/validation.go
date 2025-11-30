package toon

// Validate performs comprehensive validation on the response
// Returns ValidationError if validation fails
func (h *Handler) Validate() error {
	if h == nil {
		return &ValidationError{
			Code:    ErrCodeNilHandler,
			Message: "handler is nil",
		}
	}

	if h.resp == nil {
		return &ValidationError{
			Code:    ErrCodeNilResponse,
			Message: "response is nil",
		}
	}

	// If response indicates error, ensure error object is present
	if !h.resp.Success && h.resp.Error == nil {
		return &ValidationError{
			Code:    ErrCodeInvalidResponse,
			Message: "success is false but error object is missing",
		}
	}

	// If error object is present, validate its structure
	if h.resp.Error != nil {
		if h.resp.Error.Code == "" {
			return &ValidationError{
				Code:    ErrCodeInvalidResponse,
				Message: "error code is empty",
			}
		}
		if h.resp.Error.Message == "" {
			return &ValidationError{
				Code:    ErrCodeInvalidResponse,
				Message: "error message is empty",
			}
		}
	}

	return nil
}
