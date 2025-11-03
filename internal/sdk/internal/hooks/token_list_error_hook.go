package hooks

import (
	"bytes"
	"io"
	"net/http"
)

// TokenListErrorHook handles TokenList operations for the tokens resource.
//
// IMPORTANT LIMITATION:
// Due to the tokens API not having a GET /tokens/{id} endpoint, we must use
// GET /tokens (list all). However, the generated code takes Tokens[0] which
// would be incorrect when multiple tokens exist.
//
// Since hooks don't have access to Terraform state (to know which token ID we need),
// we cannot filter the list properly without modifying generated code.
//
// Current behavior:
// - Converts 401/403 errors to 200 OK with empty list
// - Returns ALL tokens for successful requests
// - Generated code will use the first token in the list
//
// TODO: This needs generated code modification to properly filter by ID
type TokenListErrorHook struct{}

// AfterSuccess implements the afterSuccessHook interface
func (h *TokenListErrorHook) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	// Only process if we have a response
	if res == nil {
		return res, nil
	}

	// Only handle TokenList operations
	if hookCtx.OperationID != "TokenList" {
		return res, nil
	}

	// Handle permission errors (401/403) - convert to empty list
	if res.StatusCode == 401 || res.StatusCode == 403 {
		res.StatusCode = 200
		res.Status = "200 OK"

		// Return empty token list so resource creation doesn't fail on permission errors
		emptyListBody := `{"tokens":[]}`
		res.Body = io.NopCloser(bytes.NewBufferString(emptyListBody))
		res.ContentLength = int64(len(emptyListBody))
		res.Header.Set("Content-Type", "application/json")

		return res, nil
	}

	// For successful responses, return as-is
	// Note: The generated code will incorrectly use Tokens[0] when multiple tokens exist
	// This cannot be fixed with hooks alone - requires generated code modification
	return res, nil
}
