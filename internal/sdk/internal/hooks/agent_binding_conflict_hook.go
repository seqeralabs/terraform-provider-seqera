package hooks

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

// AgentBindingConflictHook rewrites the 409 that CreateAgent returns when the
// agent's service account binding is invalid into a 400.
//
// The Platform answers 409 both for a duplicate agent name and for a service
// account that is disabled, not a workspace participant, or whose role lacks
// the agent execute permission. The generated Create reports every 409 as
// "Resource Already Exists" and suggests an import, which is wrong for the
// binding case and hides the Platform's message. As a 400 the response takes
// the generic error path, which prints the response body.
type AgentBindingConflictHook struct{}

// bindingConflictMarker prefixes every binding error from the Platform's
// AgentIdentityValidator.
const bindingConflictMarker = "The agent's service account"

// AfterSuccess implements the afterSuccessHook interface
func (h *AgentBindingConflictHook) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	if res == nil || res.StatusCode != http.StatusConflict || hookCtx.OperationID != "CreateAgent" || res.Body == nil {
		return res, nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return res, nil
	}
	res.Body = io.NopCloser(bytes.NewReader(body))

	if strings.Contains(string(body), bindingConflictMarker) {
		res.StatusCode = http.StatusBadRequest
		res.Status = "400 Bad Request"
	}
	return res, nil
}
