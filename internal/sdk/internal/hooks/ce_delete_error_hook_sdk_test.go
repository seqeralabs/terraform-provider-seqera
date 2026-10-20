package hooks_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/operations"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/shared"
)

// The generated Delete accepts 400 as a completed delete. These tests run the
// real SDK against a stub server to check that the hook makes the call fail
// instead, which is what keeps the resource in state.
func TestCEDeleteRejectionFailsSDKCall(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr bool
		wantMsg string
	}{
		{"400 fails", http.StatusBadRequest, true, "has active jobs"},
		{"409 fails", http.StatusConflict, true, "has active jobs"},
		{"404 is idempotent", http.StatusNotFound, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.status == http.StatusBadRequest || tt.status == http.StatusConflict {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.status)
					_, _ = w.Write([]byte(`{"message":"Compute environment has active jobs"}`))
					return
				}
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()

			client := sdk.New(sdk.WithServerURL(srv.URL), sdk.WithSecurity(shared.Security{BearerAuth: "token"}))
			ws := int64(1)
			res, err := client.ComputeEnvs.DeleteAzureBatchCE(context.Background(), operations.DeleteAzureBatchCERequest{ComputeEnvID: "ce1", WorkspaceID: ws})

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if res.StatusCode != tt.status {
					t.Fatalf("got status %d, want %d", res.StatusCode, tt.status)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantMsg) {
				t.Fatalf("got err %v, want it to contain %q", err, tt.wantMsg)
			}
		})
	}
}
