package toon

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHandlerWithValidResponse(t *testing.T) {
	body := []byte(`{
		"success": true,
		"data": {"id": 1, "name": "test"},
		"meta": {"request_id": "req-123"}
	}`)

	handler, err := NewHandler(body)
	require.NoError(t, err)
	require.NotNil(t, handler)
	assert.True(t, handler.IsSuccess())
	assert.False(t, handler.IsError())
	assert.Equal(t, "req-123", handler.GetRequestID())
}

func TestNewHandlerWithError(t *testing.T) {
	body := []byte(`{
		"success": false,
		"error": {
			"code": "INVALID_INPUT",
			"message": "Invalid input provided"
		}
	}`)

	handler, err := NewHandler(body)
	require.NoError(t, err)
	require.NotNil(t, handler)
	assert.False(t, handler.IsSuccess())
	assert.True(t, handler.IsError())

	errObj := handler.GetError()
	assert.NotNil(t, errObj)
	assert.Equal(t, "INVALID_INPUT", errObj.Code)
	assert.Equal(t, "Invalid input provided", errObj.Message)
}

func TestNewHandlerWithNilBody(t *testing.T) {
	handler, err := NewHandler(nil)
	assert.Error(t, err)
	assert.Nil(t, handler)

	var valErr *ValidationError
	assert.ErrorAs(t, err, &valErr)
	assert.Equal(t, ErrCodeEmptyResponse, valErr.Code)
}

func TestNewHandlerWithEmptyBody(t *testing.T) {
	handler, err := NewHandler([]byte{})
	assert.Error(t, err)
	assert.Nil(t, handler)

	var valErr *ValidationError
	assert.ErrorAs(t, err, &valErr)
	assert.Equal(t, ErrCodeEmptyResponse, valErr.Code)
}

func TestNewHandlerWithInvalidJSON(t *testing.T) {
	body := []byte(`{invalid json}`)

	handler, err := NewHandler(body)
	assert.Error(t, err)
	assert.Nil(t, handler)

	var valErr *ValidationError
	assert.ErrorAs(t, err, &valErr)
	assert.Equal(t, ErrCodeJSONUnmarshal, valErr.Code)
}

func TestFromHTTPResponseWithValidResponse(t *testing.T) {
	body := `{"success": true, "data": {"id": 1}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)

	handler, err := FromHTTPResponse(resp)
	require.NoError(t, err)
	require.NotNil(t, handler)
	assert.True(t, handler.IsSuccess())
}

func TestFromHTTPResponseWithNilResponse(t *testing.T) {
	handler, err := FromHTTPResponse(nil)
	assert.Error(t, err)
	assert.Nil(t, handler)

	var valErr *ValidationError
	assert.ErrorAs(t, err, &valErr)
	assert.Equal(t, ErrCodeInvalidResponse, valErr.Code)
}

func TestFromHTTPResponseWithStatusCodeMismatch(t *testing.T) {
	body := `{"success": true}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(body))
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)

	handler, err := FromHTTPResponse(resp)
	assert.Error(t, err)
	assert.Nil(t, handler)

	var valErr *ValidationError
	assert.ErrorAs(t, err, &valErr)
	assert.Equal(t, ErrCodeInvalidStatusCode, valErr.Code)
}

func TestUnmarshalData(t *testing.T) {
	type TestData struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	body := []byte(`{
		"success": true,
		"data": {"id": 42, "name": "test"}
	}`)

	handler, err := NewHandler(body)
	require.NoError(t, err)

	var data TestData
	err = handler.UnmarshalData(&data)
	require.NoError(t, err)
	assert.Equal(t, 42, data.ID)
	assert.Equal(t, "test", data.Name)
}

func TestUnmarshalDataWithNilTarget(t *testing.T) {
	body := []byte(`{"success": true, "data": {"id": 1}}`)
	handler, err := NewHandler(body)
	require.NoError(t, err)

	err = handler.UnmarshalData(nil)
	assert.Error(t, err)

	var valErr *ValidationError
	assert.ErrorAs(t, err, &valErr)
	assert.Equal(t, ErrCodeInvalidResponse, valErr.Code)
}

func TestUnmarshalDataWithEmptyData(t *testing.T) {
	body := []byte(`{"success": true}`)
	handler, err := NewHandler(body)
	require.NoError(t, err)

	var data map[string]interface{}
	err = handler.UnmarshalData(&data)
	assert.Error(t, err)

	var valErr *ValidationError
	assert.ErrorAs(t, err, &valErr)
	assert.Equal(t, ErrCodeEmptyData, valErr.Code)
}

func TestUnmarshalDataTypeError(t *testing.T) {
	body := []byte(`{"success": true, "data": {"id": 1}}`)
	handler, err := NewHandler(body)
	require.NoError(t, err)

	var data []string
	err = handler.UnmarshalData(&data)
	assert.Error(t, err)

	var valErr *ValidationError
	assert.ErrorAs(t, err, &valErr)
	assert.Equal(t, ErrCodeJSONUnmarshal, valErr.Code)
}

func TestRateLimit(t *testing.T) {
	resetTime := time.Now().Add(time.Hour)
	body := []byte(`{
		"success": true,
		"meta": {
			"rate_limit": {
				"limit": 1000,
				"remaining": 500,
				"reset": "` + resetTime.Format(time.RFC3339) + `"
			}
		}
	}`)

	handler, err := NewHandler(body)
	require.NoError(t, err)

	rl := handler.GetRateLimit()
	assert.NotNil(t, rl)
	assert.Equal(t, 1000, rl.Limit)
	assert.Equal(t, 500, rl.Remaining)
	assert.False(t, handler.IsRateLimited())

	body = []byte(`{
		"success": true,
		"meta": {
			"rate_limit": {
				"limit": 1000,
				"remaining": 0,
				"reset": "` + resetTime.Format(time.RFC3339) + `"
			}
		}
	}`)

	handler, err = NewHandler(body)
	require.NoError(t, err)
	assert.True(t, handler.IsRateLimited())
}

func TestErrorStringFormatting(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name:     "with all fields",
			body:     `{"success": false, "error": {"code": "ERR", "message": "msg", "details": "details", "field": "email"}}`,
			expected: "ERR | msg | details | field: email",
		},
		{
			name:     "without details",
			body:     `{"success": false, "error": {"code": "ERR", "message": "msg"}}`,
			expected: "ERR | msg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := NewHandler([]byte(tt.body))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, handler.ErrorString())
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		shouldErr bool
		errCode   ErrCode
	}{
		{
			name:      "valid success response",
			body:      `{"success": true}`,
			shouldErr: false,
		},
		{
			name:      "valid error response",
			body:      `{"success": false, "error": {"code": "ERR", "message": "msg"}}`,
			shouldErr: false,
		},
		{
			name:      "error without code",
			body:      `{"success": false, "error": {"message": "msg"}}`,
			shouldErr: true,
			errCode:   ErrCodeInvalidResponse,
		},
		{
			name:      "error without message",
			body:      `{"success": false, "error": {"code": "ERR"}}`,
			shouldErr: true,
			errCode:   ErrCodeInvalidResponse,
		},
		{
			name:      "error false but no error object",
			body:      `{"success": false}`,
			shouldErr: true,
			errCode:   ErrCodeInvalidResponse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := NewHandler([]byte(tt.body))
			require.NoError(t, err)

			err = handler.Validate()
			if tt.shouldErr {
				assert.Error(t, err)
				var valErr *ValidationError
				assert.ErrorAs(t, err, &valErr)
				assert.Equal(t, tt.errCode, valErr.Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRawBodyIsolation(t *testing.T) {
	body := []byte(`{"success": true}`)
	handler, err := NewHandler(body)
	require.NoError(t, err)

	rawBody := handler.RawBody()
	assert.NotNil(t, rawBody)

	if len(rawBody) > 0 {
		rawBody[0] = 'X'
		assert.NotEqual(t, rawBody[0], handler.RawBody()[0])
	}
}

func TestConcurrentAccess(t *testing.T) {
	body := []byte(`{
		"success": true,
		"data": {"id": 1},
		"meta": {"request_id": "req-123"}
	}`)

	handler, err := NewHandler(body)
	require.NoError(t, err)

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			assert.True(t, handler.IsSuccess())
			assert.Equal(t, "req-123", handler.GetRequestID())
			assert.NotNil(t, handler.GetData())
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		contains string
	}{
		{
			name:     "success with request id",
			body:     `{"success": true, "meta": {"request_id": "req-123"}}`,
			contains: "req-123",
		},
		{
			name:     "error",
			body:     `{"success": false, "error": {"code": "ERR", "message": "msg"}}`,
			contains: "ERR | msg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := NewHandler([]byte(tt.body))
			require.NoError(t, err)
			assert.Contains(t, handler.String(), tt.contains)
		})
	}
}

func TestGetData(t *testing.T) {
	body := []byte(`{"success": true, "data": {"id": 1}}`)
	handler, err := NewHandler(body)
	require.NoError(t, err)

	data := handler.GetData()
	assert.NotNil(t, data)

	if len(data) > 0 {
		originalFirst := data[0]
		data[0] = 'X'
		assert.NotEqual(t, data[0], handler.GetData()[0])
		assert.Equal(t, originalFirst, handler.GetData()[0])
	}
}

func BenchmarkNewHandler(b *testing.B) {
	body := []byte(`{
		"success": true,
		"data": {"id": 1, "name": "test"},
		"meta": {"request_id": "req-123"}
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewHandler(body)
	}
}

func BenchmarkUnmarshalData(b *testing.B) {
	type TestData struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	body := []byte(`{
		"success": true,
		"data": {"id": 42, "name": "test"}
	}`)

	handler, _ := NewHandler(body)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var data TestData
		_ = handler.UnmarshalData(&data)
	}
}
