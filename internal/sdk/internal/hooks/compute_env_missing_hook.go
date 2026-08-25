package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ComputeEnvMissingHook rewrites Platform's "missing compute environment" 400 responses
// into 404s on compute-environment describe and delete operations.
//
// Platform answers 400 — not 404 — when a compute environment cannot be found:
//
//	ComputeEnvServiceImpl.describe():  BadRequestException("Unknown compute environment id: $id")
//	ComputeEnvController.delete():     BadRequestException("Unknown computeEnv: $id")
//
// The provider used to accommodate that by declaring `x-speakeasy-entity-missing-codes:
// [400, 404]` on each typed CE's #read operation. Speakeasy propagates that code set to
// the entity's delete operation as well, so every typed CE's Delete tolerated a 400 —
// meaning a genuinely rejected destroy (a CE with active jobs, say) reported
// "Destruction complete", dropped the resource from state and orphaned the cloud
// resources behind it (issue #240).
//
// Normalising the status code here lets the missing-codes set shrink to [404], which
// restores an honest Delete (`case 204, 404`) while keeping Read's ability to drop a
// vanished CE from state.
//
// The rewrite is deliberately narrow: it fires only on describe/delete operations, and
// only when the response body carries one of the two "not found" messages above. Every
// other 400 — active jobs, force-delete against a non-terminal status, Seqera Compute in
// a user workspace — is left alone and surfaces as a real error. 403 and 409 are
// untouched, so auth failures and CREATING/DELETING conflicts stay visible.
type ComputeEnvMissingHook struct{}

// maxErrorBodyBytes caps how much of a response body the hook will buffer while looking
// for a "not found" message. Platform's error bodies are a few hundred bytes; anything
// beyond this is passed through untouched rather than held in memory.
const maxErrorBodyBytes = 1 << 20 // 1 MiB

// computeEnvNotFoundMessages are the (lower-cased) Platform messages that mean
// "this compute environment does not exist" despite the 400 status.
var computeEnvNotFoundMessages = []string{
	"unknown compute environment id", // ComputeEnvServiceImpl.describe()
	"unknown computeenv",             // ComputeEnvController.deleteComputeEnvironment()
}

// AfterSuccess implements the afterSuccessHook interface
func (h *ComputeEnvMissingHook) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	if res == nil || res.StatusCode != 400 {
		return res, nil
	}

	if !isComputeEnvLookupOperation(hookCtx.OperationID) {
		return res, nil
	}

	if res.Body == nil {
		return res, nil
	}

	// Buffer the body so it can be inspected and then handed back to the SDK intact.
	buf, err := io.ReadAll(io.LimitReader(res.Body, maxErrorBodyBytes+1))
	if err != nil {
		// The body is now partially consumed and unrecoverable, so returning it would
		// hand the SDK a truncated document and surface as a bogus JSON syntax error.
		// Release the connection and report the transport failure as itself.
		_ = res.Body.Close()
		return nil, fmt.Errorf("reading %s response body: %w", hookCtx.OperationID, err)
	}

	if len(buf) > maxErrorBodyBytes {
		// Implausibly large for an error body — restore the stream and pass it through.
		res.Body = struct {
			io.Reader
			io.Closer
		}{io.MultiReader(bytes.NewReader(buf), res.Body), res.Body}
		return res, nil
	}

	_ = res.Body.Close()
	res.Body = io.NopCloser(bytes.NewReader(buf))

	if !isComputeEnvNotFoundBody(buf) {
		return res, nil
	}

	// Present it as a 404. The body is left as-is: it is already an ErrorResponse, which
	// is what the 404 branch of the generated SDK expects, and it keeps Platform's own
	// message available for diagnostics.
	res.StatusCode = 404
	res.Status = "404 Not Found"

	return res, nil
}

// isComputeEnvLookupOperation reports whether the operation is a compute-environment
// describe or delete — the only two places where a "missing CE" 400 should be read as a
// 404. Create/Update/Enable/Disable/Validate deliberately fall outside: their 400s are
// genuine validation failures.
func isComputeEnvLookupOperation(operationID string) bool {
	if !strings.HasPrefix(operationID, "Describe") && !strings.HasPrefix(operationID, "Delete") {
		return false
	}
	return strings.HasSuffix(operationID, "CE") || strings.HasSuffix(operationID, "ComputeEnv")
}

// isComputeEnvNotFoundBody reports whether the body is one of Platform's "missing
// compute environment" messages.
//
// It decodes the ErrorResponse and tests only the `message` field, anchored at the start.
// Substring-matching the whole raw body would be unsafe: any 400 that merely *contains*
// one of these phrases — a user-supplied name echoed back, or a `cause`/`path` field added
// to ErrorResponse by a later Platform version — would be rewritten to 404, and on a
// delete a 404 reads as "Destruction complete" while the environment is still there. That
// is exactly the #240 failure mode this hook exists to prevent.
//
// A body that is not valid JSON, or that carries no message, is left alone.
func isComputeEnvNotFoundBody(body []byte) bool {
	var parsed struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return false
	}

	message := strings.ToLower(strings.TrimSpace(parsed.Message))
	for _, msg := range computeEnvNotFoundMessages {
		if strings.HasPrefix(message, msg) {
			return true
		}
	}
	return false
}
