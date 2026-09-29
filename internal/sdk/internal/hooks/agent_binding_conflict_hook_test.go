package hooks

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

func conflictResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

func TestAgentBindingConflictHook(t *testing.T) {
	const binding = `{"message":"The agent's service account is not a workspace participant. Add it to this workspace and try again."}`
	const duplicate = `{"message":"Duplicate agent name"}`

	cases := []struct {
		name       string
		operation  string
		status     int
		body       string
		wantStatus int
	}{
		{name: "binding conflict on create becomes bad request", operation: "CreateAgent", status: 409, body: binding, wantStatus: 400},
		{name: "duplicate name on create stays conflict", operation: "CreateAgent", status: 409, body: duplicate, wantStatus: 409},
		{name: "other operations untouched", operation: "CreateLabel", status: 409, body: binding, wantStatus: 409},
		{name: "success untouched", operation: "CreateAgent", status: 200, body: `{"agent":{}}`, wantStatus: 200},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hook := &AgentBindingConflictHook{}
			res, err := hook.AfterSuccess(AfterSuccessContext{HookContext: HookContext{OperationID: tc.operation}}, conflictResponse(tc.status, tc.body))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res.StatusCode != tc.wantStatus {
				t.Fatalf("status = %d, want %d", res.StatusCode, tc.wantStatus)
			}
			// The body must still be readable in full: the provider prints it.
			got, _ := io.ReadAll(res.Body)
			if string(got) != tc.body {
				t.Fatalf("body = %q, want %q", got, tc.body)
			}
		})
	}
}
