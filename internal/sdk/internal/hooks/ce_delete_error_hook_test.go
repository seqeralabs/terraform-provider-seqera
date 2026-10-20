package hooks

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func deleteResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body))}
}

func TestCEDeleteErrorHook(t *testing.T) {
	h := &CEDeleteErrorHook{}

	tests := []struct {
		name    string
		opID    string
		status  int
		body    string
		wantErr string
	}{
		{"409 shows message", "DeleteAzureBatchCE", 409, `{"message":"Compute environment has active jobs"}`, "HTTP 409): Compute environment has active jobs"},
		{"400 shows message", "DeleteAWSBatchCE", 400, `{"message":"boom"}`, "HTTP 400): boom"},
		{"generic delete", "DeleteComputeEnv", 409, `{"message":"busy"}`, "busy"},
		{"non-json body", "DeleteSlurmCE", 400, `oops`, "oops"},
		{"empty body", "DeleteGCPCloudCE", 409, ``, "no error message returned"},
		{"204 ok", "DeleteAzureBatchCE", 204, ``, ""},
		{"404 ok", "DeleteAzureBatchCE", 404, ``, ""},
		{"other op untouched", "DeleteCredentials", 400, `{"message":"x"}`, ""},
		{"describe untouched", "DescribeAzureBatchCE", 400, `{"message":"x"}`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := h.AfterSuccess(AfterSuccessContext{HookContext: HookContext{OperationID: tt.opID}}, deleteResponse(tt.status, tt.body))
			if res == nil {
				t.Fatal("response must be passed through")
			}
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("got error %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
