// Package computeenvs calls the compute environment operations of the Seqera Platform API v2 on behalf of the
// Terraform resources, through the client generated from the v2 OpenAPI description (package seqerav2).
//
// The resources still build and read the v1-shaped models Speakeasy generated for them; translate.go converts those
// to and from the v2 wire format, so the resources keep their schemas and state.
package computeenvs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	seqerav2 "github.com/seqeralabs/terraform-provider-seqera/internal/apiv2/client"
)

// Polling settings, unchanged from the v1 compute environment status hook.
const (
	initialWait    = 2 * time.Second
	pollInterval   = 10 * time.Second
	overallTimeout = 5 * time.Minute
)

// Doer sends HTTP requests; the SDK's HTTP client satisfies it.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client calls the v2 compute environment operations.
type Client struct {
	api *seqerav2.ClientWithResponses
}

// New returns a client for the server at serverURL (the same base URL as v1, such as `https://api.cloud.seqera.io`),
// authenticating every request with the bearer token returned by token.
func New(serverURL string, doer Doer, token func(context.Context) (string, error)) (*Client, error) {
	api, err := seqerav2.NewClientWithResponses(serverURL,
		seqerav2.WithHTTPClient(&retryingDoer{next: doer}),
		seqerav2.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			t, err := token(ctx)
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+t)
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}
	return &Client{api: api}, nil
}

// Response is the outcome of a call: the status code, the raw response (with its body readable again, so that
// errors can print the problem document), and, for the operations that answer with a compute environment, that
// compute environment as a v1 describe response body.
type Response struct {
	StatusCode  int
	RawResponse *http.Response
	// Describe holds the compute environment as a v1 describe response body, when the call returned one.
	Describe []byte
	// ComputeEnvID is the identifier of the compute environment created, for Create.
	ComputeEnvID string
}

func wsp(workspaceID int64) string { return strconv.FormatInt(workspaceID, 10) }

func newResponse(raw *http.Response, body []byte) *Response {
	raw.Body = io.NopCloser(bytes.NewReader(body))
	return &Response{StatusCode: raw.StatusCode, RawResponse: raw}
}

// Create creates the compute environment described by the v1 create request body, and waits until it is AVAILABLE,
// as the v1 status hook did. On success the response carries the compute environment as last read.
func (c *Client) Create(ctx context.Context, workspaceID int64, v1Body []byte) (*Response, error) {
	body, err := toV2WriteModel(v1Body)
	if err != nil {
		return nil, err
	}
	res, err := c.api.CreateComputeEnvWithBodyWithResponse(ctx, wsp(workspaceID), nil, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	out := newResponse(res.HTTPResponse, res.Body)
	if res.JSON201 == nil || res.JSON201.Id == nil {
		return out, nil
	}
	out.ComputeEnvID = *res.JSON201.Id

	current := res.Body
	if res.JSON201.Status == nil || *res.JSON201.Status != seqerav2.AVAILABLE {
		current, err = c.waitUntilAvailable(ctx, workspaceID, out.ComputeEnvID)
		if err != nil {
			return nil, fmt.Errorf("failed to poll compute environment status: %w", err)
		}
	}
	if out.Describe, err = toV1Describe(current, workspaceID); err != nil {
		return nil, err
	}
	return out, nil
}

// Get reads the compute environment. include names the v1 attributes to add (`labels`, `resources`).
func (c *Client) Get(ctx context.Context, workspaceID int64, id string, include []string) (*Response, error) {
	var params *seqerav2.GetComputeEnvParams
	if fields := v2Include(include); len(fields) > 0 {
		params = &seqerav2.GetComputeEnvParams{Include: &fields}
	}
	res, err := c.api.GetComputeEnvWithResponse(ctx, wsp(workspaceID), id, params)
	if err != nil {
		return nil, err
	}
	out := newResponse(res.HTTPResponse, res.Body)
	if res.JSON200 != nil {
		if out.Describe, err = toV1Describe(res.Body, workspaceID); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// Update applies the v1 update request body. On success the response carries the updated compute environment.
func (c *Client) Update(ctx context.Context, workspaceID int64, id string, v1Body []byte) (*Response, error) {
	body, err := toV2Patch(v1Body)
	if err != nil {
		return nil, err
	}
	res, err := c.api.UpdateComputeEnvWithBodyWithResponse(ctx, wsp(workspaceID), id, nil, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	out := newResponse(res.HTTPResponse, res.Body)
	if res.JSON200 != nil {
		if out.Describe, err = toV1Describe(res.Body, workspaceID); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// Delete deletes the compute environment, and waits until it is gone, as the v1 status hook did.
func (c *Client) Delete(ctx context.Context, workspaceID int64, id string, force *bool) (*Response, error) {
	var params *seqerav2.DeleteComputeEnvParams
	if force != nil {
		params = &seqerav2.DeleteComputeEnvParams{Force: force}
	}
	res, err := c.api.DeleteComputeEnvWithResponse(ctx, wsp(workspaceID), id, params)
	if err != nil {
		return nil, err
	}
	out := newResponse(res.HTTPResponse, res.Body)
	if res.StatusCode() == http.StatusNoContent {
		if err := c.waitUntilGone(ctx, workspaceID, id); err != nil {
			return nil, fmt.Errorf("failed to poll compute environment status: %w", err)
		}
	}
	return out, nil
}

// Enable enables the compute environment.
func (c *Client) Enable(ctx context.Context, workspaceID int64, id string) (*Response, error) {
	res, err := c.api.EnableComputeEnvWithResponse(ctx, wsp(workspaceID), id)
	if err != nil {
		return nil, err
	}
	return newResponse(res.HTTPResponse, res.Body), nil
}

// Disable disables the compute environment.
func (c *Client) Disable(ctx context.Context, workspaceID int64, id string) (*Response, error) {
	res, err := c.api.DisableComputeEnvWithResponse(ctx, wsp(workspaceID), id)
	if err != nil {
		return nil, err
	}
	return newResponse(res.HTTPResponse, res.Body), nil
}

// SetPrimary makes the compute environment the primary one of the workspace.
func (c *Client) SetPrimary(ctx context.Context, workspaceID int64, id string) (*Response, error) {
	res, err := c.api.SetPrimaryComputeEnvWithResponse(ctx, wsp(workspaceID), id)
	if err != nil {
		return nil, err
	}
	return newResponse(res.HTTPResponse, res.Body), nil
}

// waitUntilAvailable polls the compute environment until it is AVAILABLE, failing when it turns ERRORED or INVALID.
// KEPT: no server equivalent. v2 creates compute environments asynchronously and offers no way to wait for the
// outcome, so the client keeps polling, with the settings of the v1 status hook.
func (c *Client) waitUntilAvailable(ctx context.Context, workspaceID int64, id string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, overallTimeout)
	defer cancel()
	last := ""
	for wait := initialWait; ; wait = pollInterval {
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for compute environment (last status: %s): %w", last, ctx.Err())
		}
		res, err := c.api.GetComputeEnvWithResponse(ctx, wsp(workspaceID), id, nil)
		if err != nil {
			return nil, err
		}
		if res.JSON200 == nil {
			return nil, fmt.Errorf("describe returned status %d: %s", res.StatusCode(), string(res.Body))
		}
		if res.JSON200.Status != nil {
			last = string(*res.JSON200.Status)
		}
		switch last {
		case string(seqerav2.AVAILABLE):
			return res.Body, nil
		case string(seqerav2.ERRORED), string(seqerav2.INVALID):
			return nil, fmt.Errorf("compute environment entered error state: %s. The errored compute environment must be manually deleted from the Seqera Platform before Terraform can recreate it", last)
		}
	}
}

// waitUntilGone polls the compute environment until reading it is refused, which v2 does for a deleted compute
// environment (403) or an unknown one (404).
// KEPT: no server equivalent. v2 deletes compute environments asynchronously and offers no way to wait for the
// outcome, so the client keeps polling, with the settings of the v1 status hook.
func (c *Client) waitUntilGone(ctx context.Context, workspaceID int64, id string) error {
	ctx, cancel := context.WithTimeout(ctx, overallTimeout)
	defer cancel()
	last := ""
	for wait := initialWait; ; wait = pollInterval {
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for compute environment (last status: %s): %w", last, ctx.Err())
		}
		res, err := c.api.GetComputeEnvWithResponse(ctx, wsp(workspaceID), id, nil)
		if err != nil {
			return err
		}
		switch res.StatusCode() {
		case http.StatusForbidden, http.StatusNotFound:
			return nil
		case http.StatusOK:
			if res.JSON200 != nil && res.JSON200.Status != nil {
				last = string(*res.JSON200.Status)
			}
		default:
			return fmt.Errorf("describe returned status %d: %s", res.StatusCode(), string(res.Body))
		}
	}
}

// v2Include maps the v1 describe attributes to the v2 include fields.
func v2Include(attributes []string) []string {
	var out []string
	for _, a := range attributes {
		switch a {
		case "labels":
			out = append(out, "resourceLabels")
		case "resources":
			out = append(out, "resources")
		}
	}
	return out
}

// MarshalV1 is json.Marshal, named for the call sites that hand v1 models to this package.
func MarshalV1(v any) ([]byte, error) { return json.Marshal(v) }

// DecodeV1 decodes a v1 describe response body into the v1 model the resource reads its state from.
func DecodeV1[T any](body []byte) (*T, error) {
	if body == nil {
		return nil, nil
	}
	var out T
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
