// Integration coverage for ComputeEnvMissingHook (issue #240).
//
// The #240 bug was not in any single layer — it was an interlock between the hook chain,
// the generated SDK's status-code switch, and the resource's Delete tolerance. These tests
// drive a real SDK client against a stub Platform so the status code the resource layer
// would actually observe is asserted end to end.
//
// Hand-maintained; protected by .genignore.
package sdk_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sdk "github.com/seqeralabs/terraform-provider-seqera/internal/sdk"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/operations"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/shared"
)

// stubPlatform serves a single status/body pair for every request.
func stubPlatform(t *testing.T, status int, body string) *sdk.Seqera {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	return sdk.New(
		sdk.WithServerURL(srv.URL),
		sdk.WithSecurity(shared.Security{BearerAuth: "test-token"}),
	)
}

// stubSuccessfulDelete serves 204 for the DELETE and 404 for the describe poll that
// ComputeEnvStatusHook fires afterwards (the hook reads a 404 there as "gone, success").
// It records the DELETE query string so parameter threading can be asserted.
func stubSuccessfulDelete(t *testing.T, gotQuery *string) *sdk.Seqera {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			*gotQuery = r.URL.RawQuery
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not found"}`))
	}))
	t.Cleanup(srv.Close)

	return sdk.New(sdk.WithServerURL(srv.URL), sdk.WithSecurity(shared.Security{BearerAuth: "t"}))
}

const (
	describeNotFound = `{"message":"Unknown compute environment id: 3xampleId"}`
	deleteNotFound   = `{"message":"Unknown computeEnv: 3xampleId"}`
	activeJobs       = `{"message":"Compute environment '3xampleId' has active jobs"}`
)

// Deleting a CE that Platform says does not exist must present as 404, which the
// resource's `case 204, 404` tolerates — so an idempotent destroy still converges even if
// refresh did not run first.
func TestDeleteMissingComputeEnvSurfacesAs404(t *testing.T) {
	client := stubPlatform(t, 400, deleteNotFound)

	res, err := client.ComputeEnvs.DeleteSlurmCE(context.Background(), operations.DeleteSlurmCERequest{
		ComputeEnvID: "3xampleId",
		WorkspaceID:  1,
	})
	if err != nil {
		t.Fatalf("DeleteSlurmCE returned error: %v", err)
	}
	if res.StatusCode != 404 {
		t.Errorf("got status %d, want 404", res.StatusCode)
	}
}

// The #240 regression guard. A delete Platform genuinely rejected must reach the resource
// layer as a 400 so it falls into Delete's `default` branch and raises a diagnostic,
// rather than being reported as "Destruction complete".
func TestRejectedDeleteSurfacesAs400(t *testing.T) {
	client := stubPlatform(t, 400, activeJobs)

	res, err := client.ComputeEnvs.DeleteSlurmCE(context.Background(), operations.DeleteSlurmCERequest{
		ComputeEnvID: "3xampleId",
		WorkspaceID:  1,
	})
	if err != nil {
		t.Fatalf("DeleteSlurmCE returned error: %v", err)
	}
	if res.StatusCode != 400 {
		t.Fatalf("got status %d, want 400 — a rejected delete must not look like success", res.StatusCode)
	}
	// The rejection reason must survive for the diagnostic.
	if res.ErrorResponse == nil {
		t.Fatal("ErrorResponse not parsed; the diagnostic would lose the rejection reason")
	}
	if got := res.ErrorResponse.Message; got != "Compute environment '3xampleId' has active jobs" {
		t.Errorf("got message %q, want the active-jobs rejection", got)
	}
}

// A CREATING or already-DELETING CE answers 409. The hook only touches 400, so this path
// is untouched by the #240 fix — but the typed CE deletes now *declare* 409, so it arrives
// as a typed conflict the resource reports through its `default` branch, rather than as the
// SDK's "unknown status code returned".
func TestConflictingDeleteSurfacesAs409(t *testing.T) {
	client := stubPlatform(t, 409, `{"message":"Deletion for compute environment already in progress"}`)

	res, err := client.ComputeEnvs.DeleteSlurmCE(context.Background(), operations.DeleteSlurmCERequest{
		ComputeEnvID: "3xampleId",
		WorkspaceID:  1,
	})
	if err != nil {
		t.Fatalf("DeleteSlurmCE returned error: %v", err)
	}
	if res.StatusCode != 409 {
		t.Fatalf("got status %d, want 409", res.StatusCode)
	}
	if res.ErrorResponse == nil {
		t.Fatal("ErrorResponse not parsed; the diagnostic would lose the conflict reason")
	}
	if !strings.Contains(res.ErrorResponse.Message, "already in progress") {
		t.Errorf("got message %q, want the in-progress conflict", res.ErrorResponse.Message)
	}
}

// Every CE delete must type 409 rather than falling through to "unknown status code".
func TestAllTypedComputeEnvDeletesType409(t *testing.T) {
	for name, call := range deleteCalls(context.Background()) {
		t.Run(name, func(t *testing.T) {
			status, err := call(stubPlatform(t, 409, `{"message":"Deletion for compute environment already in progress"}`))
			if err != nil {
				t.Fatalf("delete returned error rather than a typed 409: %v", err)
			}
			if status != 409 {
				t.Errorf("got status %d, want 409", status)
			}
		})
	}
}

// `force=true` must reach Platform as a query parameter — without it, deleting an ERRORED
// CE returns 204 and does nothing, so destroy hangs in the status hook until it times out.
func TestForceIsSentAsQueryParam(t *testing.T) {
	var gotQuery string
	client := stubSuccessfulDelete(t, &gotQuery)

	res, err := client.ComputeEnvs.DeleteSlurmCE(context.Background(), operations.DeleteSlurmCERequest{
		ComputeEnvID: "3xampleId",
		WorkspaceID:  1,
		Force:        sdk.Pointer(true),
	})
	if err != nil {
		t.Fatalf("DeleteSlurmCE returned error: %v", err)
	}
	if res.StatusCode != 204 {
		t.Errorf("got status %d, want 204", res.StatusCode)
	}
	if !strings.Contains(gotQuery, "force=true") {
		t.Errorf("force not sent; query was %q", gotQuery)
	}
}

// An explicit false must reach Platform as false, rather than being treated as unset
// or accidentally enabling force-delete.
func TestForceFalseIsSentAsQueryParam(t *testing.T) {
	var gotQuery string
	client := stubSuccessfulDelete(t, &gotQuery)

	res, err := client.ComputeEnvs.DeleteSlurmCE(context.Background(), operations.DeleteSlurmCERequest{
		ComputeEnvID: "3xampleId",
		WorkspaceID:  1,
		Force:        sdk.Pointer(false),
	})
	if err != nil {
		t.Fatalf("DeleteSlurmCE returned error: %v", err)
	}
	if res.StatusCode != 204 {
		t.Errorf("got status %d, want 204", res.StatusCode)
	}
	if !strings.Contains(gotQuery, "force=false") {
		t.Errorf("force=false not sent; query was %q", gotQuery)
	}
	if strings.Contains(gotQuery, "force=true") {
		t.Errorf("force=true sent for an explicit false value; query was %q", gotQuery)
	}
}

// An unset force must not reach the wire at all, so Platform keeps its default
// (non-force) behaviour. The parameter carries `example: false` rather than
// `default: false` in the spec: a schema default would make the attribute Computed and
// show `+ force = false` against every pre-existing CE, forcing an in-place Update on
// each — including wedged ones, which are precisely the environments force exists for.
//
// Above all, an unset force must never reach the wire as true: that would turn every
// destroy into a force-delete, skipping forge cleanup and orphaning cloud resources.
func TestForceOmittedWhenUnset(t *testing.T) {
	var gotQuery string
	client := stubSuccessfulDelete(t, &gotQuery)

	if _, err := client.ComputeEnvs.DeleteSlurmCE(context.Background(), operations.DeleteSlurmCERequest{
		ComputeEnvID: "3xampleId",
		WorkspaceID:  1,
	}); err != nil {
		t.Fatalf("DeleteSlurmCE returned error: %v", err)
	}
	if strings.Contains(gotQuery, "force") {
		t.Errorf("force sent despite being unset; query was %q", gotQuery)
	}
}

// Misusing force (a CE that is not ERRORED/INVALID/DELETING) is a 400 that must surface as
// an error, not be swallowed as a missing CE.
func TestForceMisuseSurfacesAs400(t *testing.T) {
	client := stubPlatform(t, 400, `{"message":"Force-delete is only available for compute environments in ERRORED, INVALID, or DELETING status"}`)

	res, err := client.ComputeEnvs.DeleteSlurmCE(context.Background(), operations.DeleteSlurmCERequest{
		ComputeEnvID: "3xampleId",
		WorkspaceID:  1,
		Force:        sdk.Pointer(true),
	})
	if err != nil {
		t.Fatalf("DeleteSlurmCE returned error: %v", err)
	}
	if res.StatusCode != 400 {
		t.Errorf("got status %d, want 400", res.StatusCode)
	}
}

// The regression guard for #201: describing a purged CE must present as 404 so Read drops
// it from state instead of erroring.
func TestDescribeMissingComputeEnvSurfacesAs404(t *testing.T) {
	client := stubPlatform(t, 400, describeNotFound)

	res, err := client.ComputeEnvs.DescribeSlurmCE(context.Background(), operations.DescribeSlurmCERequest{
		ComputeEnvID: "3xampleId",
		WorkspaceID:  1,
	})
	if err != nil {
		t.Fatalf("DescribeSlurmCE returned error: %v", err)
	}
	if res.StatusCode != 404 {
		t.Errorf("got status %d, want 404", res.StatusCode)
	}
}

// A describe rejected for some other reason must still surface as a 400 error rather than
// silently removing a live resource from state.
func TestDescribeOtherBadRequestStays400(t *testing.T) {
	client := stubPlatform(t, 400, `{"message":"Something else went wrong"}`)

	res, err := client.ComputeEnvs.DescribeSlurmCE(context.Background(), operations.DescribeSlurmCERequest{
		ComputeEnvID: "3xampleId",
		WorkspaceID:  1,
	})
	if err != nil {
		t.Fatalf("DescribeSlurmCE returned error: %v", err)
	}
	if res.StatusCode != 400 {
		t.Errorf("got status %d, want 400", res.StatusCode)
	}
}

// deleteCalls returns one delete invocation per CE resource type, so behaviour can be
// asserted across all ten rather than just the one used in the focused tests above.
func deleteCalls(ctx context.Context) map[string]func(*sdk.Seqera) (int, error) {
	return map[string]func(*sdk.Seqera) (int, error){
		"AWSBatchCE": func(c *sdk.Seqera) (int, error) {
			r, err := c.ComputeEnvs.DeleteAWSBatchCE(ctx, operations.DeleteAWSBatchCERequest{ComputeEnvID: "x", WorkspaceID: 1})
			return statusOf(r, err)
		},
		"AwsCloudCE": func(c *sdk.Seqera) (int, error) {
			r, err := c.ComputeEnvs.DeleteAwsCloudCE(ctx, operations.DeleteAwsCloudCERequest{ComputeEnvID: "x", WorkspaceID: 1})
			return statusOf(r, err)
		},
		"AzureBatchCE": func(c *sdk.Seqera) (int, error) {
			r, err := c.ComputeEnvs.DeleteAzureBatchCE(ctx, operations.DeleteAzureBatchCERequest{ComputeEnvID: "x", WorkspaceID: 1})
			return statusOf(r, err)
		},
		"AzureCloudCE": func(c *sdk.Seqera) (int, error) {
			r, err := c.ComputeEnvs.DeleteAzureCloudCE(ctx, operations.DeleteAzureCloudCERequest{ComputeEnvID: "x", WorkspaceID: 1})
			return statusOf(r, err)
		},
		"GCPBatchCE": func(c *sdk.Seqera) (int, error) {
			r, err := c.ComputeEnvs.DeleteGCPBatchCE(ctx, operations.DeleteGCPBatchCERequest{ComputeEnvID: "x", WorkspaceID: 1})
			return statusOf(r, err)
		},
		"GCPCloudCE": func(c *sdk.Seqera) (int, error) {
			r, err := c.ComputeEnvs.DeleteGCPCloudCE(ctx, operations.DeleteGCPCloudCERequest{ComputeEnvID: "x", WorkspaceID: 1})
			return statusOf(r, err)
		},
		"ManagedComputeCE": func(c *sdk.Seqera) (int, error) {
			r, err := c.ComputeEnvs.DeleteManagedComputeCE(ctx, operations.DeleteManagedComputeCERequest{ComputeEnvID: "x", WorkspaceID: 1})
			return statusOf(r, err)
		},
		"SlurmCE": func(c *sdk.Seqera) (int, error) {
			r, err := c.ComputeEnvs.DeleteSlurmCE(ctx, operations.DeleteSlurmCERequest{ComputeEnvID: "x", WorkspaceID: 1})
			return statusOf(r, err)
		},
		"ComputeEnv": func(c *sdk.Seqera) (int, error) {
			r, err := c.ComputeEnvs.DeleteComputeEnv(ctx, operations.DeleteComputeEnvRequest{ComputeEnvID: "x", WorkspaceID: sdk.Pointer(int64(1))})
			return statusOf(r, err)
		},
		"AWSComputeEnv": func(c *sdk.Seqera) (int, error) {
			r, err := c.ComputeEnvs.DeleteAWSComputeEnv(ctx, operations.DeleteAWSComputeEnvRequest{ComputeEnvID: "x", WorkspaceID: 1})
			return statusOf(r, err)
		},
	}
}

// The behaviour must hold for every CE resource type, not just the one used above — the
// bug affected all of them.
func TestAllTypedComputeEnvDeletesNormaliseMissing(t *testing.T) {
	for name, call := range deleteCalls(context.Background()) {
		t.Run(name+"/missing", func(t *testing.T) {
			status, err := call(stubPlatform(t, 400, deleteNotFound))
			if err != nil {
				t.Fatalf("delete returned error: %v", err)
			}
			if status != 404 {
				t.Errorf("got status %d, want 404", status)
			}
		})
		t.Run(name+"/rejected", func(t *testing.T) {
			status, err := call(stubPlatform(t, 400, activeJobs))
			if err != nil {
				t.Fatalf("delete returned error: %v", err)
			}
			if status != 400 {
				t.Errorf("got status %d, want 400 — a rejected delete must not look like success", status)
			}
		})
	}
}

// statusOf pulls the status code out of any of the delete response types.
func statusOf(res interface{ GetStatusCode() int }, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	return res.GetStatusCode(), nil
}
