package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/shared"
)

/*
Compute Environment Status Hook

This is a global SDK hook injected into the Terraform provider, filtered by operation ID.
It handles asynchronous compute environment operations by polling for completion status.

For Compute Environment Creation:
  - The API responds with a 200 status code containing the computeEnvId
  - We poll the describe endpoint until the status field becomes "AVAILABLE"
  - If status becomes "ERRORED" or "INVALID", the operation fails
  - Polling configuration: up to 60 attempts with exponential backoff (1s for first 3 attempts, then 5s)
  - Total timeout: approximately 5 minutes

For Compute Environment Deletion:
  - The API responds with a 204 status code acknowledging the deletion request
  - We poll the describe endpoint until either:
    * The resource returns 404 (not found), or
    * The deleted field in the response is true
  - Note: Earlier versions of Seqera have no DELETING status phase - the resource goes
    directly from existing to 404/deleted
  - Same polling configuration as creation operations

This hook ensures Terraform operations are synchronous, preventing state inconsistencies
caused by the API's asynchronous behavior.
*/

const (
	// Operation IDs for compute environment create operations
	opCreateAWSBatchCE        = "CreateAWSBatchCE"
	opCreateAWSComputeEnv     = "CreateAWSComputeEnv"
	opCreateManagedComputeCE  = "CreateManagedComputeCE"
	opCreateGenericComputeEnv = "CreateComputeEnv"

	// Operation IDs for compute environment delete operations
	opDeleteAWSBatchCE        = "DeleteAWSBatchCE"
	opDeleteAWSComputeEnv     = "DeleteAWSComputeEnv"
	opDeleteManagedComputeCE  = "DeleteManagedComputeCE"
	opDeleteGenericComputeEnv = "DeleteComputeEnv"

	// ComputeEnvMaxPollAttempts defines maximum polling attempts (5 minutes total with default interval)
	ComputeEnvMaxPollAttempts = 60
	// ComputeEnvPollInterval defines time between polling attempts
	ComputeEnvPollInterval = 5 * time.Second
	// ComputeEnvInitialWait defines initial wait before first poll (gives API time to initialize)
	ComputeEnvInitialWait = 2 * time.Second
	// ComputeEnvHTTPTimeout defines timeout for individual HTTP requests
	ComputeEnvHTTPTimeout = 30 * time.Second
)

// ComputeEnvStatusHook polls compute environment status:
// - For create operations: polls until AVAILABLE
// - For delete operations: polls until resource is deleted (deleted: true in API response)
type ComputeEnvStatusHook struct{}

// AfterSuccess implements the afterSuccessHook interface
func (h *ComputeEnvStatusHook) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	// Only process if we have a response
	if res == nil {
		return res, nil
	}

	// Only process successful responses for polling
	// 200 for create operations, 204 for delete operations
	if res.StatusCode != 200 && res.StatusCode != 204 {
		return res, nil
	}

	// Check if this is a compute environment create or delete operation
	createOps := []string{opCreateAWSBatchCE, opCreateAWSComputeEnv, opCreateManagedComputeCE, opCreateGenericComputeEnv}
	deleteOps := []string{opDeleteAWSBatchCE, opDeleteAWSComputeEnv, opDeleteManagedComputeCE, opDeleteGenericComputeEnv}

	isCreateOperation := slices.Contains(createOps, hookCtx.OperationID)
	isDeleteOperation := slices.Contains(deleteOps, hookCtx.OperationID)

	if !isCreateOperation && !isDeleteOperation {
		return res, nil
	}

	var computeEnvID string
	var bodyBytes []byte
	var err error

	// For delete operations, extract compute env ID from request path
	if isDeleteOperation {
		computeEnvID, err = extractComputeEnvIDFromPath(res.Request)
		if err != nil {
			return res, fmt.Errorf("failed to extract computeEnvId from path: %w", err)
		}
	} else {
		// For create operations, extract from response body
		bodyBytes, err = io.ReadAll(res.Body)
		if err != nil {
			return res, fmt.Errorf("failed to read response body: %w", err)
		}

		// Restore the body for the SDK to read
		res.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse the response to get the computeEnvId
		var createResponse struct {
			ComputeEnvID string `json:"computeEnvId"`
		}

		if err := json.Unmarshal(bodyBytes, &createResponse); err != nil {
			return res, fmt.Errorf("failed to parse create response: %w", err)
		}

		if createResponse.ComputeEnvID == "" {
			return res, fmt.Errorf("computeEnvId not found in create response")
		}

		computeEnvID = createResponse.ComputeEnvID
	}

	// Extract workspaceId from the request URL query parameters
	workspaceID, err := extractWorkspaceID(res.Request)
	if err != nil {
		return res, fmt.Errorf("failed to extract workspaceId: %w", err)
	}

	// Poll for status
	finalStatus, err := h.pollComputeEnvStatus(
		hookCtx.Context,
		hookCtx.BaseURL,
		computeEnvID,
		workspaceID,
		res.Request.Header.Get("Authorization"),
		isDeleteOperation,
	)
	if err != nil {
		return res, fmt.Errorf("failed to poll compute environment status: %w", err)
	}

	// For create operations, verify status is AVAILABLE
	if isCreateOperation && finalStatus != string(shared.ComputeEnvStatusAvailable) {
		return res, fmt.Errorf("compute environment creation failed, final status: %s", finalStatus)
	}

	// For delete operations, finalStatus will be empty (resource deleted)
	return res, nil
}

// extractWorkspaceID extracts the workspaceId from the request URL
func extractWorkspaceID(req *http.Request) (string, error) {
	if req == nil || req.URL == nil {
		return "", fmt.Errorf("request or URL is nil")
	}

	workspaceID := req.URL.Query().Get("workspaceId")
	if workspaceID == "" {
		return "", fmt.Errorf("workspaceId not found in query parameters")
	}

	return workspaceID, nil
}

// extractComputeEnvIDFromPath extracts the computeEnvId from the request path
// Path format: /api/compute-envs/{computeEnvId} or /compute-envs/{computeEnvId}
func extractComputeEnvIDFromPath(req *http.Request) (string, error) {
	if req == nil || req.URL == nil {
		return "", fmt.Errorf("request or URL is nil")
	}

	path := req.URL.Path
	// Expected format: /api/compute-envs/{computeEnvId} or /compute-envs/{computeEnvId}
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Find the index of "compute-envs" in the path parts
	computeEnvsIndex := -1
	for i, part := range parts {
		if part == "compute-envs" {
			computeEnvsIndex = i
			break
		}
	}

	if computeEnvsIndex == -1 || computeEnvsIndex+1 >= len(parts) {
		return "", fmt.Errorf("invalid path format: %s", path)
	}

	computeEnvID := parts[computeEnvsIndex+1]
	if computeEnvID == "" {
		return "", fmt.Errorf("computeEnvId not found in path: %s", path)
	}

	// Validate computeEnvID doesn't contain invalid characters
	if strings.Contains(computeEnvID, "?") || strings.Contains(computeEnvID, "/") {
		return "", fmt.Errorf("invalid computeEnvId extracted: %s", computeEnvID)
	}

	return computeEnvID, nil
}

// pollComputeEnvStatus polls the compute environment status until it's AVAILABLE (for create) or deleted flag is true (for delete)
func (h *ComputeEnvStatusHook) pollComputeEnvStatus(
	ctx context.Context,
	baseURL string,
	computeEnvID string,
	workspaceID string,
	authHeader string,
	isDeleteOperation bool,
) (string, error) {
	// Wait a bit before starting to poll - gives the API time to initialize
	time.Sleep(ComputeEnvInitialWait)

	// Create HTTP client once for all polling attempts
	client := &http.Client{
		Timeout: ComputeEnvHTTPTimeout,
	}

	describeURL := fmt.Sprintf("%s/compute-envs/%s?workspaceId=%s",
		strings.TrimSuffix(baseURL, "/"),
		computeEnvID,
		workspaceID,
	)

	for attempt := 1; attempt <= ComputeEnvMaxPollAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("polling cancelled: %w", ctx.Err())
		default:
		}

		req, err := http.NewRequestWithContext(ctx, "GET", describeURL, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create describe request: %w", err)
		}

		req.Header.Set("Authorization", authHeader)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to describe compute environment: %w", err)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		closeErr := resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("failed to read describe response: %w", err)
		}
		if closeErr != nil {
			return "", fmt.Errorf("failed to close response body: %w", closeErr)
		}

		// For delete operations, a 404 means the resource is deleted (success)
		if isDeleteOperation && resp.StatusCode == 404 {
			return "", nil
		}

		if resp.StatusCode != 200 {
			return "", fmt.Errorf("describe returned status %d: %s", resp.StatusCode, string(bodyBytes))
		}

		// Parse the response to get status and deleted flag
		var describeResponse struct {
			ComputeEnv struct {
				Status  string `json:"status"`
				Deleted bool   `json:"deleted"`
			} `json:"computeEnv"`
		}

		if err := json.Unmarshal(bodyBytes, &describeResponse); err != nil {
			return "", fmt.Errorf("failed to parse describe response: %w", err)
		}

		status := describeResponse.ComputeEnv.Status
		deleted := describeResponse.ComputeEnv.Deleted

		// Handle delete operations
		if isDeleteOperation {
			if deleted {
				return "", nil // Successfully deleted
			}
			// Continue polling until 404 or deleted flag is true
			// Note: Earlier Seqera versions have no DELETING phase
		} else {
			// Handle create operations
			if status == string(shared.ComputeEnvStatusAvailable) {
				return status, nil
			}
			// Check for error states
			if status == string(shared.ComputeEnvStatusErrored) || status == string(shared.ComputeEnvStatusInvalid) {
				return status, fmt.Errorf("compute environment entered error state: %s. The errored compute environment must be manually deleted from the Seqera Platform before Terraform can recreate it", status)
			}
		}

		// Wait before next attempt with exponential backoff for initial polls
		if attempt < ComputeEnvMaxPollAttempts {
			interval := ComputeEnvPollInterval
			// Poll more frequently in the first few attempts
			if attempt <= 3 {
				interval = time.Second
			}
			time.Sleep(interval)
		}
	}

	if isDeleteOperation {
		return "", fmt.Errorf("timeout waiting for compute environment to be deleted after %d attempts", ComputeEnvMaxPollAttempts)
	}
	return "", fmt.Errorf("timeout waiting for compute environment to become available after %d attempts", ComputeEnvMaxPollAttempts)
}
