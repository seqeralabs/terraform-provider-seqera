package computeenvs

import (
	"context"
	"errors"
	"math"
	"net/http"
	"time"
)

// retryingDoer retries transient failures with the policy the v1 SDK applies to every call (overlays/retries.yaml):
// network errors and HTTP 429, 502, 503 and 504, with an exponential backoff from 500ms to 30s, for 5 minutes at most.
// KEPT: no server equivalent. Transient failures are a property of the network path, not something the v2 contract
// can take over.
type retryingDoer struct {
	next Doer
}

const (
	retryInitialInterval = 500 * time.Millisecond
	retryMaxInterval     = 30 * time.Second
	retryMaxElapsed      = 5 * time.Minute
	retryExponent        = 1.5
)

func (d *retryingDoer) Do(req *http.Request) (*http.Response, error) {
	start := time.Now()
	for attempt := 0; ; attempt++ {
		if attempt > 0 && req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			req.Body = body
		}
		res, err := d.next.Do(req)
		if !retryable(res, err) || req.Context().Err() != nil {
			return res, err
		}
		wait := time.Duration(float64(retryInitialInterval) * math.Pow(retryExponent, float64(attempt)))
		if wait > retryMaxInterval {
			wait = retryMaxInterval
		}
		if time.Since(start)+wait > retryMaxElapsed {
			return res, err
		}
		if res != nil && res.Body != nil {
			_ = res.Body.Close()
		}
		select {
		case <-time.After(wait):
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	}
}

func retryable(res *http.Response, err error) bool {
	if err != nil {
		return !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
	}
	switch res.StatusCode {
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	}
	return false
}
