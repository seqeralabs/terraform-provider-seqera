package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// CEDeleteErrorHook turns a rejected compute environment deletion into an error
// that carries the Platform's message.
//
// The generated Delete of the per-type compute environment resources derives its
// tolerated status codes from the Describe operation (400 and 404, so a purged
// environment drops out of state), which makes a 400 on delete look like a
// successful destroy. The Platform also answers 409 when it refuses a deletion,
// for example while the environment has active jobs. Neither means the
// environment is gone, so both are surfaced as errors and the resource stays in
// state. 404 is left alone so deletes remain idempotent.
type CEDeleteErrorHook struct{}

// AfterSuccess implements the afterSuccessHook interface
func (h *CEDeleteErrorHook) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	if res == nil || (res.StatusCode != http.StatusBadRequest && res.StatusCode != http.StatusConflict) {
		return res, nil
	}

	opID := hookCtx.OperationID
	isCEDelete := strings.HasPrefix(opID, "Delete") &&
		(strings.Contains(opID, "ComputeEnv") || strings.HasSuffix(opID, "CE"))
	if !isCEDelete {
		return res, nil
	}

	var body []byte
	if res.Body != nil {
		body, _ = io.ReadAll(res.Body)
		res.Body = io.NopCloser(bytes.NewReader(body))
	}

	return res, fmt.Errorf("compute environment deletion rejected by Seqera Platform (HTTP %d): %s", res.StatusCode, errorMessage(body))
}

// errorMessage extracts the "message" field of a Platform ErrorResponse, falling
// back to the raw body.
func errorMessage(body []byte) string {
	var parsed struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Message != "" {
		return parsed.Message
	}
	if msg := strings.TrimSpace(string(body)); msg != "" {
		return msg
	}
	return "no error message returned"
}
