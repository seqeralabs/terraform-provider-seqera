package common

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// unexpectedStatusDetail formats a detail string for an unexpected HTTP
// status response. Includes the full request/response dump from
// DebugResponse so users see what the server actually said. `fmt.Sprintf`
// is unavoidable here because the action verb is interpolated.
func unexpectedStatusDetail(action string, res *http.Response) string {
	return fmt.Sprintf("Status %d while %s:\n%s", res.StatusCode, action, DebugResponse(res))
}

// UnexpectedStatusErr returns an error for an unexpected HTTP status,
// suitable for return values in helper closures and worker functions.
// The error message includes the request/response dump from DebugResponse.
func UnexpectedStatusErr(action string, res *http.Response) error {
	return errors.New(unexpectedStatusDetail(action, res))
}

// AddUnexpectedStatus appends an "Unexpected API response" error diagnostic
// containing the full request/response dump. Use this at every
// non-success-status site in resource/data source methods so error messages
// stay consistent and never drop the wire-level detail.
func AddUnexpectedStatus(diags *diag.Diagnostics, action string, res *http.Response) {
	diags.AddError("Unexpected API response", unexpectedStatusDetail(action, res))
}
