package hooks

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

// GenericResourceErrorHook transforms 403 responses to 404 for all describe operations
// to allow Terraform to properly handle deleted resources.
// When a resource is deleted in the web UI, the API returns 403 instead of 404.
// This hook converts 403 responses to 404 for all describe/read operations so that
// Terraform removes the resource from state instead of showing an error.
type GenericResourceErrorHook struct{}

// AfterSuccess implements the afterSuccessHook interface
func (h *GenericResourceErrorHook) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	// Only process if we have a response
	if res == nil {
		return res, nil
	}

	// Check if this is a 403 response
	if res.StatusCode != 403 {
		return res, nil
	}

	// Check if this is a describe/read operation
	if !strings.HasPrefix(hookCtx.OperationID, "Describe") {
		return res, nil
	}

	// Convert 403 to 404 so Terraform treats it as "not found"
	// Create a valid JSON error response body for the SDK to parse
	res.StatusCode = 404
	res.Status = "404 Not Found"

	// Create a valid JSON error response body
	errorBody := `{"message":"Resource not found or has been deleted"}`
	res.Body = io.NopCloser(bytes.NewBufferString(errorBody))
	res.ContentLength = int64(len(errorBody))
	res.Header.Set("Content-Type", "application/json")

	return res, nil
}
