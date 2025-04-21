package haproxy

import (
	"context"
	"testing"
	"time"
)

// TestHAProxyError tests the error handling for HAProxy responses
func TestHAProxyError(t *testing.T) {
	// Test creating and formatting HAProxy errors
	err := NewHAProxyError(3, "Server not found", "show servers")
	if err.Error() != "[3]: Server not found (command: show servers)" {
		t.Errorf("Unexpected error message format: %s", err.Error())
	}

	if err.Code != 3 {
		t.Errorf("Expected code 3, got %d", err.Code)
	}

	if err.Message != "Server not found" {
		t.Errorf("Expected message 'Server not found', got '%s'", err.Message)
	}

	if err.Command != "show servers" {
		t.Errorf("Expected command 'show servers', got '%s'", err.Command)
	}
}

// TestContextWithTimeout tests the context timeout handling
func TestContextWithTimeout(t *testing.T) {
	// Create a context with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Sleep to make the context timeout
	time.Sleep(100 * time.Millisecond)

	// Check if the context is done
	if ctx.Err() == nil {
		t.Error("Expected context to be done, but it wasn't")
	}

	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded error, got: %v", ctx.Err())
	}
}

// TestProcessHAProxyResponse tests the processing of HAProxy responses
func TestProcessHAProxyResponse(t *testing.T) {
	// Test cases with different HAProxy response formats
	testCases := []struct {
		name     string
		response string
		success  bool
		errCode  int
	}{
		{
			name:     "Successful response",
			response: "Successfully executed command",
			success:  true,
		},
		{
			name:     "Error code 0 (general error)",
			response: "[0]: Unknown command",
			success:  false,
			errCode:  0,
		},
		{
			name:     "Error code 1 (startup/reload error)",
			response: "[1]: Cannot reload during startup",
			success:  false,
			errCode:  1,
		},
		{
			name:     "Error code 2 (resource shortage)",
			response: "[2]: Out of memory",
			success:  false,
			errCode:  2,
		},
		{
			name:     "Error code 3 (not found)",
			response: "[3]: No such server",
			success:  false,
			errCode:  3,
		},
	}

	// Helper function that simulates the response handling logic
	processResponse := func(response, command string) (string, error) {
		if len(response) > 4 {
			if response[0] == '[' && response[2] == ']' && response[3] == ':' {
				code := int(response[1] - '0')
				message := response[4:]
				return "", NewHAProxyError(code, message, command)
			}
		}
		return response, nil
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := processResponse(tc.response, "test command")

			if tc.success {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
				}
				if result != tc.response {
					t.Errorf("Expected response '%s', got '%s'", tc.response, result)
				}
			} else {
				if err == nil {
					t.Error("Expected error but got success")
				}

				haErr, ok := err.(HAProxyError)
				if !ok {
					t.Errorf("Expected HAProxyError but got %T", err)
				} else if haErr.Code != tc.errCode {
					t.Errorf("Expected error code %d, got %d", tc.errCode, haErr.Code)
				}
			}
		})
	}
}
