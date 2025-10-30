package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ComputeEnvStatusHook polls compute environment status until AVAILABLE after creation operations
type ComputeEnvStatusHook struct{}

// AfterSuccess implements the afterSuccessHook interface
func (h *ComputeEnvStatusHook) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	// Only process if we have a response
	if res == nil {
		return res, nil
	}

	// Only process successful creation responses for polling
	if res.StatusCode != 200 {
		return res, nil
	}

	// Check if this is a compute environment create operation
	isAWSBatchCECreate := hookCtx.OperationID == "CreateAWSBatchCE"
	isAWSComputeEnvCreate := hookCtx.OperationID == "CreateAWSComputeEnv"

	if !isAWSBatchCECreate && !isAWSComputeEnvCreate {
		return res, nil
	}

	// Read the response body to extract the compute environment ID
	bodyBytes, err := io.ReadAll(res.Body)
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

	// Extract workspaceId from the request URL query parameters
	workspaceID, err := extractWorkspaceID(res.Request)
	if err != nil {
		return res, fmt.Errorf("failed to extract workspaceId: %w", err)
	}

	// Poll for status
	finalStatus, err := h.pollComputeEnvStatus(
		hookCtx.Context,
		hookCtx.BaseURL,
		createResponse.ComputeEnvID,
		workspaceID,
		res.Request.Header.Get("Authorization"),
	)
	if err != nil {
		return res, fmt.Errorf("failed to poll compute environment status: %w", err)
	}

	// If status is not AVAILABLE, return an error
	if finalStatus != "AVAILABLE" {
		return res, fmt.Errorf("compute environment creation failed, final status: %s", finalStatus)
	}

	// Return the original create response - status will be fetched on first read
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

// pollComputeEnvStatus polls the compute environment status until it's AVAILABLE
func (h *ComputeEnvStatusHook) pollComputeEnvStatus(
	ctx context.Context,
	baseURL string,
	computeEnvID string,
	workspaceID string,
	authHeader string,
) (string, error) {
	const (
		maxAttempts  = 60              // Maximum number of polling attempts
		pollInterval = 5 * time.Second // Time between polling attempts
		initialWait  = 2 * time.Second // Initial wait before first poll
	)

	// Wait a bit before starting to poll - gives the API time to initialize
	time.Sleep(initialWait)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	describeURL := fmt.Sprintf("%s/compute-envs/%s?workspaceId=%s",
		strings.TrimSuffix(baseURL, "/"),
		computeEnvID,
		workspaceID,
	)

	for attempt := 1; attempt <= maxAttempts; attempt++ {
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
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("failed to read describe response: %w", err)
		}

		if resp.StatusCode != 200 {
			return "", fmt.Errorf("describe returned status %d: %s", resp.StatusCode, string(bodyBytes))
		}

		// Parse the response to get status
		var describeResponse struct {
			ComputeEnv struct {
				Status string `json:"status"`
			} `json:"computeEnv"`
		}

		if err := json.Unmarshal(bodyBytes, &describeResponse); err != nil {
			return "", fmt.Errorf("failed to parse describe response: %w", err)
		}

		status := describeResponse.ComputeEnv.Status

		// Check if we've reached AVAILABLE status
		if status == "AVAILABLE" {
			return status, nil
		}

		// Check for error states
		if status == "ERRORED" || status == "INVALID" {
			return status, fmt.Errorf("compute environment entered error state: %s. The errored compute environment must be manually deleted from the Seqera Platform before Terraform can recreate it", status)
		}

		// Wait before next attempt
		if attempt < maxAttempts {
			time.Sleep(pollInterval)
		}
	}

	return "", fmt.Errorf("timeout waiting for compute environment to become available after %d attempts", maxAttempts)
}
