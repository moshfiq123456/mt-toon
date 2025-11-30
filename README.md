# mt-toon

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Production-grade Go package for handling Toon API responses with comprehensive error handling, type-safe data access, and concurrent-safe operations.

## Features

- ✅ Type-safe response handling
- ✅ Comprehensive error handling with `ValidationError`
- ✅ Thread-safe concurrent access
- ✅ Rate limit tracking and management
- ✅ Request ID correlation for debugging
- ✅ Zero external dependencies (except tests)
- ✅ Full test coverage (20+ test cases)
- ✅ Production-ready error codes
- ✅ Nil-safety throughout

## Installation

\`\`\`bash
go get github.com/moshfiq123456/mt-toon
\`\`\`

## Quick Start

\`\`\`go
package main

import (
	"fmt"
	"log"
	"github.com/moshfiq123456/mt-toon/pkg/toon"
)

func main() {
	body := []byte(\`{"success": true, "data": {"id": 1}}\`)
	
	handler, err := toon.NewHandler(body)
	if err != nil {
		log.Fatal(err)
	}
	
	if handler.IsSuccess() {
		fmt.Println("✓ Request successful")
	}
}
\`\`\`

## Usage Examples

### Handling Success Responses

\`\`\`go
handler, err := toon.NewHandler(body)
if err != nil {
	log.Fatal(err)
}

if handler.IsSuccess() {
	fmt.Printf("Request ID: %s\n", handler.GetRequestID())
}
\`\`\`

### Handling Errors

\`\`\`go
if handler.IsError() {
	fmt.Printf("Error: %s\n", handler.ErrorString())
}
\`\`\`

### Type-Safe Unmarshaling

\`\`\`go
type User struct {
	ID   int    \`json:"id"\`
	Name string \`json:"name"\`
}

var user User
if err := handler.UnmarshalData(&user); err != nil {
	log.Fatal(err)
}
\`\`\`

### Rate Limit Management

\`\`\`go
if handler.IsRateLimited() {
	fmt.Println("Rate limited!")
}
fmt.Println(handler.GetRateLimitStatus())
\`\`\`

## Testing

\`\`\`bash
make test
make test-coverage
make test-verbose
\`\`\`

## Development

\`\`\`bash
make build
make fmt
make lint
make example
\`\`\`

## License

MIT License - see [LICENSE](LICENSE)

## Author

Moshfiq - [GitHub](https://github.com/moshfiq123456)
