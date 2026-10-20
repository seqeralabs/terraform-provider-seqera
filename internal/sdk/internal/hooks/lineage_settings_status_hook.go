package hooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

/*
Lineage Settings Status Hook

Updating workspace lineage settings (PUT, 204 no body) with AUTO provisioning is
asynchronous: the API status is PENDING until it becomes AVAILABLE or ERRORED
(with a message). This hook polls the GET settings endpoint after a successful
UpdateLineageWorkspaceSettings call so Terraform only continues once provisioning
has finished. In MANUAL mode the status is AVAILABLE immediately.
*/

const (
	lineageSettingsOperationID = "UpdateLineageWorkspaceSettings"

	defaultLineageInitialWait    = 2 * time.Second
	defaultLineagePollInterval   = 10 * time.Second
	defaultLineageRetryInterval  = 1 * time.Second
	defaultLineageOverallTimeout = 10 * time.Minute
	lineageHTTPTimeout           = 30 * time.Second
)

// LineageSettingsStatusHook polls lineage settings status until AVAILABLE.
// Zero-value duration fields fall back to defaults.
type LineageSettingsStatusHook struct {
	InitialWait    time.Duration
	PollInterval   time.Duration
	RetryInterval  time.Duration
	OverallTimeout time.Duration
}

func durationOrDefault(d, def time.Duration) time.Duration {
	if d <= 0 {
		return def
	}
	return d
}

// AfterSuccess implements the afterSuccessHook interface
func (h *LineageSettingsStatusHook) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	if res == nil || hookCtx.OperationID != lineageSettingsOperationID || res.StatusCode != 204 {
		return res, nil
	}

	orgID, workspaceID, err := extractLineageIDsFromPath(res.Request)
	if err != nil {
		return res, fmt.Errorf("failed to extract ids from path: %w", err)
	}

	ctx := hookCtx.Context
	if ctx == nil {
		ctx = context.Background()
	}

	if err := h.pollLineageStatus(ctx, hookCtx.BaseURL, orgID, workspaceID, res.Request.Header.Get("Authorization")); err != nil {
		return res, fmt.Errorf("failed to poll lineage settings status: %w", err)
	}
	return res, nil
}

// extractLineageIDsFromPath parses /orgs/{orgId}/workspaces/{workspaceId}/settings/lineage
func extractLineageIDsFromPath(req *http.Request) (string, string, error) {
	if req == nil || req.URL == nil {
		return "", "", fmt.Errorf("request or URL is nil")
	}
	parts := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	var orgID, workspaceID string
	for i := 0; i+1 < len(parts); i++ {
		switch parts[i] {
		case "orgs":
			orgID = parts[i+1]
		case "workspaces":
			workspaceID = parts[i+1]
		}
	}
	if orgID == "" || workspaceID == "" {
		return "", "", fmt.Errorf("invalid path format: %s", req.URL.Path)
	}
	return orgID, workspaceID, nil
}

func sleepCtx(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return true
	case <-ctx.Done():
		return false
	}
}

func (h *LineageSettingsStatusHook) pollLineageStatus(ctx context.Context, baseURL, orgID, workspaceID, authHeader string) error {
	if authHeader == "" {
		return fmt.Errorf("authorization header is empty")
	}

	initialWait := durationOrDefault(h.InitialWait, defaultLineageInitialWait)
	pollInterval := durationOrDefault(h.PollInterval, defaultLineagePollInterval)
	retryInterval := durationOrDefault(h.RetryInterval, defaultLineageRetryInterval)
	overall := durationOrDefault(h.OverallTimeout, defaultLineageOverallTimeout)

	ctx, cancel := context.WithTimeout(ctx, overall)
	defer cancel()

	if !sleepCtx(ctx, initialWait) {
		return fmt.Errorf("polling cancelled during initial wait: %w", ctx.Err())
	}

	client := &http.Client{Timeout: lineageHTTPTimeout}
	url := fmt.Sprintf("%s/orgs/%s/workspaces/%s/settings/lineage", strings.TrimSuffix(baseURL, "/"), orgID, workspaceID)

	var lastStatus string
	for {
		if ctx.Err() != nil {
			return fmt.Errorf("timeout waiting for lineage settings provisioning (last status: %s): %w", lastStatus, ctx.Err())
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", authHeader)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("request cancelled (last status: %s): %w", lastStatus, err)
			}
			lastStatus = fmt.Sprintf("error: %v", err)
			if !sleepCtx(ctx, retryInterval) {
				return fmt.Errorf("timeout during retry (last status: %s): %w", lastStatus, ctx.Err())
			}
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("failed to read response: %w", readErr)
		}

		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			lastStatus = fmt.Sprintf("status %d", resp.StatusCode)
			if !sleepCtx(ctx, retryInterval) {
				return fmt.Errorf("timeout during retry (last status: %s): %w", lastStatus, ctx.Err())
			}
			continue
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("lineage settings returned status %d: %s", resp.StatusCode, string(body))
		}

		var parsed struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(body, &parsed); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
		lastStatus = parsed.Status

		switch parsed.Status {
		case "AVAILABLE":
			return nil
		case "ERRORED":
			return fmt.Errorf("lineage provisioning failed: %s", parsed.Message)
		}

		if !sleepCtx(ctx, pollInterval) {
			return fmt.Errorf("timeout waiting for lineage settings provisioning (last status: %s): %w", lastStatus, ctx.Err())
		}
	}
}
