package hooks

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// newResponse builds a minimal 400-style response carrying the given body.
func newResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode:    status,
		Status:        http.StatusText(status),
		Body:          io.NopCloser(strings.NewReader(body)),
		ContentLength: int64(len(body)),
		Header:        http.Header{"Content-Type": []string{"application/json"}},
	}
}

func runHook(t *testing.T, operationID string, res *http.Response) *http.Response {
	t.Helper()
	h := &ComputeEnvMissingHook{}
	out, err := h.AfterSuccess(AfterSuccessContext{HookContext: HookContext{OperationID: operationID}}, res)
	if err != nil {
		t.Fatalf("AfterSuccess returned error: %v", err)
	}
	return out
}

func readBody(t *testing.T, res *http.Response) string {
	t.Helper()
	if res.Body == nil {
		return ""
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	return string(b)
}

// The describe message must be rewritten to a 404 so Read drops the CE from state.
func TestDescribeUnknownComputeEnvBecomes404(t *testing.T) {
	for _, opID := range []string{
		"DescribeSlurmCE",
		"DescribeAWSBatchCE",
		"DescribeAwsCloudCE",
		"DescribeManagedComputeCE",
		"DescribeComputeEnv",
		"DescribeAWSComputeEnv",
	} {
		body := `{"message":"Unknown compute environment id: 3xampleId"}`
		res := runHook(t, opID, newResponse(400, body))
		if res.StatusCode != 404 {
			t.Errorf("%s: got status %d, want 404", opID, res.StatusCode)
		}
		// The body must survive intact — the SDK still parses it, and it carries
		// Platform's message for diagnostics.
		if got := readBody(t, res); got != body {
			t.Errorf("%s: body altered: got %q, want %q", opID, got, body)
		}
	}
}

// The delete endpoint uses different wording for the same condition.
func TestDeleteUnknownComputeEnvBecomes404(t *testing.T) {
	for _, opID := range []string{"DeleteSlurmCE", "DeleteGCPCloudCE", "DeleteComputeEnv"} {
		res := runHook(t, opID, newResponse(400, `{"message":"Unknown computeEnv: 3xampleId"}`))
		if res.StatusCode != 404 {
			t.Errorf("%s: got status %d, want 404", opID, res.StatusCode)
		}
	}
}

// This is the #240 regression guard: a delete Platform actually rejected must stay a 400
// so the resource surfaces an error instead of reporting "Destruction complete".
func TestRejectedDeleteStays400(t *testing.T) {
	rejections := []string{
		`{"message":"Compute environment 'abc123' has active jobs"}`,
		`{"message":"Force-delete is only available for compute environments in ERRORED, INVALID, or DELETING status"}`,
		`{"message":"Cannot delete Seqera Compute in user workspaces"}`,
		`{"message":"Oops, something went wrong"}`,
	}
	for _, body := range rejections {
		res := runHook(t, "DeleteSlurmCE", newResponse(400, body))
		if res.StatusCode != 400 {
			t.Errorf("body %q: got status %d, want 400", body, res.StatusCode)
		}
		if got := readBody(t, res); got != body {
			t.Errorf("body altered: got %q, want %q", got, body)
		}
	}
}

// Create/Update/Enable/Disable/Validate 400s are genuine validation failures and must not
// be reinterpreted, even if they happen to mention an unknown compute environment.
func TestNonLookupOperationsUntouched(t *testing.T) {
	for _, opID := range []string{
		"CreateSlurmCE",
		"UpdateSlurmCE",
		"UpdateComputeEnv",
		"EnableComputeEnv",
		"DisableComputeEnv",
		"ValidateComputeEnv",
	} {
		res := runHook(t, opID, newResponse(400, `{"message":"Unknown compute environment id: x"}`))
		if res.StatusCode != 400 {
			t.Errorf("%s: got status %d, want 400", opID, res.StatusCode)
		}
	}
}

// Unrelated entities must not be caught by the operation-ID matching.
func TestUnrelatedOperationsUntouched(t *testing.T) {
	for _, opID := range []string{"DescribeCredentials", "DeletePipeline", "DescribeWorkflow"} {
		res := runHook(t, opID, newResponse(400, `{"message":"Unknown compute environment id: x"}`))
		if res.StatusCode != 400 {
			t.Errorf("%s: got status %d, want 400", opID, res.StatusCode)
		}
	}
}

// Only 400 is in scope: 403 belongs to GenericResourceErrorHook, and 409 (CREATING /
// already-DELETING) must keep surfacing as a conflict.
func TestOtherStatusCodesUntouched(t *testing.T) {
	for _, status := range []int{200, 204, 403, 409, 500} {
		res := runHook(t, "DeleteSlurmCE", newResponse(status, `{"message":"Unknown computeEnv: x"}`))
		if res.StatusCode != status {
			t.Errorf("got status %d, want %d", res.StatusCode, status)
		}
	}
}

func TestNilResponseAndNilBody(t *testing.T) {
	if got := runHook(t, "DeleteSlurmCE", nil); got != nil {
		t.Errorf("nil response: got %v, want nil", got)
	}

	res := &http.Response{StatusCode: 400, Header: http.Header{}}
	if got := runHook(t, "DeleteSlurmCE", res); got.StatusCode != 400 {
		t.Errorf("nil body: got status %d, want 400", got.StatusCode)
	}
}

// An implausibly large body is passed through with the stream intact rather than buffered.
func TestOversizedBodyPassedThrough(t *testing.T) {
	body := strings.Repeat("x", maxErrorBodyBytes+64)
	res := runHook(t, "DeleteSlurmCE", newResponse(400, body))
	if res.StatusCode != 400 {
		t.Errorf("got status %d, want 400", res.StatusCode)
	}
	if got := readBody(t, res); got != body {
		t.Errorf("body truncated: got %d bytes, want %d", len(got), len(body))
	}
}

// Matching is case-insensitive, since the wording differs between the two endpoints.
func TestMessageMatchingIsCaseInsensitive(t *testing.T) {
	for _, body := range []string{
		`{"message":"UNKNOWN COMPUTE ENVIRONMENT ID: x"}`,
		`{"message":"unknown computeenv: x"}`,
	} {
		res := runHook(t, "DeleteSlurmCE", newResponse(400, body))
		if res.StatusCode != 404 {
			t.Errorf("body %q: got status %d, want 404", body, res.StatusCode)
		}
	}
}

// The hook must not leave the body half-consumed for the next reader in the chain.
func TestBodyReusableAfterRewrite(t *testing.T) {
	body := `{"message":"Unknown compute environment id: x"}`
	res := runHook(t, "DescribeSlurmCE", newResponse(400, body))
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, res.Body); err != nil {
		t.Fatalf("copying body: %v", err)
	}
	if buf.String() != body {
		t.Errorf("got %q, want %q", buf.String(), body)
	}
}

// Finding 4 regression guard: a 400 that merely *contains* a not-found phrase — echoed
// back in a user-supplied field, or in a field a later Platform version adds — must not be
// rewritten. On a delete, a wrongly-produced 404 reads as "Destruction complete" while the
// environment is still there, which is the #240 failure mode itself.
func TestNotFoundPhraseElsewhereInBodyIsIgnored(t *testing.T) {
	bodies := []string{
		// The phrase echoed inside a different field.
		`{"message":"Compute environment 'abc' has active jobs","path":"/unknown computeenv"}`,
		// A CE whose *name* happens to contain the phrase.
		`{"message":"Compute environment 'unknown compute environment id' has active jobs"}`,
		// The phrase present, but not as the message.
		`{"message":"Bad request","cause":"Unknown compute environment id: xyz"}`,
	}
	for _, body := range bodies {
		res := runHook(t, "DeleteSlurmCE", newResponse(400, body))
		if res.StatusCode != 400 {
			t.Errorf("body %q was wrongly rewritten to %d; want 400", body, res.StatusCode)
		}
	}
}

// Only the message field, anchored at the start, triggers the rewrite.
func TestMessageMustBeAnchored(t *testing.T) {
	// Prefixed with other text: not Platform's message, so leave it alone.
	res := runHook(t, "DeleteSlurmCE", newResponse(400, `{"message":"Warning: Unknown computeEnv: x"}`))
	if res.StatusCode != 400 {
		t.Errorf("non-anchored message rewritten to %d; want 400", res.StatusCode)
	}
	// Anchored, with the id trailing as Platform sends it.
	res = runHook(t, "DeleteSlurmCE", newResponse(400, `{"message":"Unknown computeEnv: x"}`))
	if res.StatusCode != 404 {
		t.Errorf("anchored message not rewritten; got %d, want 404", res.StatusCode)
	}
}

// A body that is not JSON, or carries no message, must be passed through untouched.
func TestNonJSONAndMessagelessBodiesIgnored(t *testing.T) {
	for _, body := range []string{
		`Unknown computeEnv: x`,             // plain text, not JSON
		`{"error":"Unknown computeEnv: x"}`, // no message field
		`{}`,
		``,
	} {
		res := runHook(t, "DeleteSlurmCE", newResponse(400, body))
		if res.StatusCode != 400 {
			t.Errorf("body %q rewritten to %d; want 400", body, res.StatusCode)
		}
	}
}

// Finding 1 regression guard: when the body cannot be read, the hook must report the
// transport failure rather than handing the SDK a silently truncated document (which
// surfaces as a bogus JSON syntax error on a destroy).
func TestBodyReadErrorIsReported(t *testing.T) {
	closed := false
	res := &http.Response{
		StatusCode: 400,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body: readCloser{
			reader: io.MultiReader(strings.NewReader(`{"message":"Unknown com`), errReader{}),
			onClose: func() error {
				closed = true
				return nil
			},
		},
	}

	h := &ComputeEnvMissingHook{}
	out, err := h.AfterSuccess(AfterSuccessContext{HookContext: HookContext{OperationID: "DeleteSlurmCE"}}, res)
	if err == nil {
		t.Fatal("expected the read failure to be reported, got nil error")
	}
	if !strings.Contains(err.Error(), "DeleteSlurmCE") {
		t.Errorf("error should name the operation, got: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil response alongside the error, got %v", out)
	}
	if !closed {
		t.Error("body was not closed; the connection would leak")
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("connection reset") }

type readCloser struct {
	reader  io.Reader
	onClose func() error
}

func (r readCloser) Read(p []byte) (int, error) { return r.reader.Read(p) }
func (r readCloser) Close() error               { return r.onClose() }
