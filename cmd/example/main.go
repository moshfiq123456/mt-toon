package main

import (
	"fmt"
	"log"

	"github.com/moshfiq123456/mt-toon/pkg/toon"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	fmt.Println("=== mt-toon Package Examples ===")
	fmt.Println()

	// Example 1: Success Response
	fmt.Println("1. Handling Success Response")
	fmt.Println("----------------------------------------")
	successBody := []byte(`{
		"success": true,
		"data": {
			"id": 1,
			"name": "John Doe",
			"email": "john@example.com"
		},
		"meta": {
			"request_id": "req-abc123",
			"api_version": "v1"
		}
	}`)

	handler, err := toon.NewHandler(successBody)
	if err != nil {
		log.Fatal(err)
	}

	if handler.IsSuccess() {
		var user User
		if err := handler.UnmarshalData(&user); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("✓ Success!\n")
		fmt.Printf("  User: %s (%s)\n", user.Name, user.Email)
		fmt.Printf("  Request ID: %s\n", handler.GetRequestID())
		fmt.Printf("  API Version: %s\n", handler.GetAPIVersion())
		fmt.Println()
	}

	// Example 2: Error Response
	fmt.Println("2. Handling Error Response")
	fmt.Println("----------------------------------------")
	errorBody := []byte(`{
		"success": false,
		"error": {
			"code": "INVALID_EMAIL",
			"message": "Email format is invalid",
			"details": "Must contain @ symbol",
			"field": "email"
		}
	}`)

	handler, err = toon.NewHandler(errorBody)
	if err != nil {
		log.Fatal(err)
	}

	if handler.IsError() {
		fmt.Printf("✗ Error Occurred!\n")
		fmt.Printf("  Details: %s\n", handler.ErrorString())
		fmt.Println()
	}

	// Example 3: Rate Limit Information
	fmt.Println("3. Rate Limit Management")
	fmt.Println("----------------------------------------")
	rateLimitBody := []byte(`{
		"success": true,
		"meta": {
			"rate_limit": {
				"limit": 1000,
				"remaining": 250,
				"reset": "2025-12-31T23:59:59Z"
			}
		}
	}`)

	handler, err = toon.NewHandler(rateLimitBody)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Rate Limit Status:\n")
	fmt.Printf("  %s\n", handler.GetRateLimitStatus())
	if handler.IsRateLimited() {
		fmt.Printf("  ⚠ Rate limited!\n")
	} else {
		fmt.Printf("  ✓ Not rate limited\n")
	}
	fmt.Println()

	// Example 4: Concurrent Access
	fmt.Println("4. Safe Concurrent Access")
	fmt.Println("----------------------------------------")
	concurrentBody := []byte(`{
		"success": true,
		"data": {"id": 42},
		"meta": {"request_id": "concurrent-test"}
	}`)

	handler, err = toon.NewHandler(concurrentBody)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool, 3)

	go func() {
		fmt.Printf("Goroutine 1: Success = %v\n", handler.IsSuccess())
		done <- true
	}()

	go func() {
		fmt.Printf("Goroutine 2: RequestID = %s\n", handler.GetRequestID())
		done <- true
	}()

	go func() {
		data := handler.GetData()
		fmt.Printf("Goroutine 3: Data size = %d bytes\n", len(data))
		done <- true
	}()

	for i := 0; i < 3; i++ {
		<-done
	}

	fmt.Println()
	fmt.Printf("✓ All examples completed successfully!\n")
}