package hooks

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// LineageSettingsNotConfiguredHook rewrites the response of FindLineageWorkspaceSettings
// to a 404 when the workspace has no lineage configuration.
// The backend answers 200 with empty fields (no credentialsId) for such workspaces,
// so without this hook Terraform would keep an empty resource in state.
// Configs in PENDING or ERRORED still carry a credentialsId and pass through untouched.
type LineageSettingsNotConfiguredHook struct{}

// AfterSuccess implements the afterSuccessHook interface
func (h *LineageSettingsNotConfiguredHook) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	if res == nil || res.StatusCode != 200 || hookCtx.OperationID != "FindLineageWorkspaceSettings" {
		return res, nil
	}

	raw, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return res, err
	}
	// Always restore the body so the SDK can read it
	res.Body = io.NopCloser(bytes.NewReader(raw))

	var parsed struct {
		CredentialsID string `json:"credentialsId"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		// Unparseable body: pass through untouched
		return res, nil
	}
	if parsed.CredentialsID != "" {
		return res, nil
	}

	errorBody := `{"message":"Lineage is not configured for this workspace"}`
	res.StatusCode = 404
	res.Status = "404 Not Found"
	res.Body = io.NopCloser(bytes.NewBufferString(errorBody))
	res.ContentLength = int64(len(errorBody))
	if res.Header == nil {
		res.Header = http.Header{}
	}
	res.Header.Set("Content-Type", "application/json")

	return res, nil
}
