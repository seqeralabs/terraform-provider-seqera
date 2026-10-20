package hooks

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func notConfiguredCall(t *testing.T, opID string, status int, body string) (*http.Response, string) {
	t.Helper()
	res := &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
	hookCtx := AfterSuccessContext{HookContext: HookContext{OperationID: opID}}
	out, err := (&LineageSettingsNotConfiguredHook{}).AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := io.ReadAll(out.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}
	return out, string(b)
}

func TestLineageSettingsNotConfiguredHook_EmptyBecomes404(t *testing.T) {
	out, body := notConfiguredCall(t, "FindLineageWorkspaceSettings", 200, `{"orgId":1,"wspId":2,"provisioned":false}`)
	if out.StatusCode != 404 {
		t.Fatalf("status = %d, want 404", out.StatusCode)
	}
	want := `{"message":"Lineage is not configured for this workspace"}`
	if body != want {
		t.Errorf("body = %q, want %q", body, want)
	}
	if ct := out.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q", ct)
	}
	if out.ContentLength != int64(len(want)) {
		t.Errorf("content length = %d", out.ContentLength)
	}
}

func TestLineageSettingsNotConfiguredHook_EmptyCredentialsBecomes404(t *testing.T) {
	out, _ := notConfiguredCall(t, "FindLineageWorkspaceSettings", 200, `{"credentialsId":""}`)
	if out.StatusCode != 404 {
		t.Fatalf("status = %d, want 404", out.StatusCode)
	}
}

func TestLineageSettingsNotConfiguredHook_ConfiguredPassesThrough(t *testing.T) {
	in := `{"credentialsId":"abc","status":"AVAILABLE"}`
	out, body := notConfiguredCall(t, "FindLineageWorkspaceSettings", 200, in)
	if out.StatusCode != 200 || body != in {
		t.Errorf("status=%d body=%q", out.StatusCode, body)
	}
}

func TestLineageSettingsNotConfiguredHook_PendingWithCredentialsPassesThrough(t *testing.T) {
	in := `{"credentialsId":"abc","status":"PENDING"}`
	out, body := notConfiguredCall(t, "FindLineageWorkspaceSettings", 200, in)
	if out.StatusCode != 200 || body != in {
		t.Errorf("status=%d body=%q", out.StatusCode, body)
	}
}

func TestLineageSettingsNotConfiguredHook_InvalidJSONPassesThrough(t *testing.T) {
	in := `not json`
	out, body := notConfiguredCall(t, "FindLineageWorkspaceSettings", 200, in)
	if out.StatusCode != 200 || body != in {
		t.Errorf("status=%d body=%q", out.StatusCode, body)
	}
}

func TestLineageSettingsNotConfiguredHook_OtherOperationsUntouched(t *testing.T) {
	in := `{"orgId":1}`
	out, body := notConfiguredCall(t, "ListWorkspaces", 200, in)
	if out.StatusCode != 200 || body != in {
		t.Errorf("status=%d body=%q", out.StatusCode, body)
	}
}
