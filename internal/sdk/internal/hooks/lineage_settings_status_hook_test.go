package hooks

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const lineageTestPath = "/orgs/7/workspaces/9/settings/lineage"

func newLineageHook() *LineageSettingsStatusHook {
	return &LineageSettingsStatusHook{
		InitialWait:    time.Millisecond,
		PollInterval:   5 * time.Millisecond,
		RetryInterval:  5 * time.Millisecond,
		OverallTimeout: 200 * time.Millisecond,
	}
}

func lineageCall(t *testing.T, h *LineageSettingsStatusHook, ctx context.Context, baseURL, opID string) error {
	t.Helper()
	req, err := http.NewRequest("PUT", baseURL+lineageTestPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer x")
	res := &http.Response{StatusCode: 204, Request: req}
	hookCtx := AfterSuccessContext{HookContext: HookContext{BaseURL: baseURL, Context: ctx, OperationID: opID}}
	_, err = h.AfterSuccess(hookCtx, res)
	return err
}

func TestLineageSettingsStatusHook_AvailableImmediately(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != lineageTestPath || r.Header.Get("Authorization") != "Bearer x" {
			t.Errorf("unexpected request %s %s auth=%q", r.Method, r.URL.Path, r.Header.Get("Authorization"))
		}
		_, _ = fmt.Fprint(w, `{"status":"AVAILABLE"}`)
	}))
	defer srv.Close()
	if err := lineageCall(t, newLineageHook(), context.Background(), srv.URL, "UpdateLineageWorkspaceSettings"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLineageSettingsStatusHook_PendingThenAvailable(t *testing.T) {
	var n int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&n, 1) <= 2 {
			_, _ = fmt.Fprint(w, `{"status":"PENDING"}`)
			return
		}
		_, _ = fmt.Fprint(w, `{"status":"AVAILABLE"}`)
	}))
	defer srv.Close()
	if err := lineageCall(t, newLineageHook(), context.Background(), srv.URL, "UpdateLineageWorkspaceSettings"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&n); got != 3 {
		t.Errorf("expected 3 polls, got %d", got)
	}
}

func TestLineageSettingsStatusHook_ErroredReturnsMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"status":"ERRORED","message":"boom"}`)
	}))
	defer srv.Close()
	err := lineageCall(t, newLineageHook(), context.Background(), srv.URL, "UpdateLineageWorkspaceSettings")
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected error containing boom, got %v", err)
	}
}

func TestLineageSettingsStatusHook_TimesOutWhilePending(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"status":"PENDING"}`)
	}))
	defer srv.Close()
	h := newLineageHook()
	h.OverallTimeout = 50 * time.Millisecond
	err := lineageCall(t, h, context.Background(), srv.URL, "UpdateLineageWorkspaceSettings")
	if err == nil || !strings.Contains(err.Error(), "PENDING") {
		t.Fatalf("expected timeout error mentioning PENDING, got %v", err)
	}
}

func TestLineageSettingsStatusHook_RetriesOn503ThenSucceeds(t *testing.T) {
	var n int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&n, 1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, `{"status":"AVAILABLE"}`)
	}))
	defer srv.Close()
	if err := lineageCall(t, newLineageHook(), context.Background(), srv.URL, "UpdateLineageWorkspaceSettings"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&n); got != 2 {
		t.Errorf("expected 2 requests, got %d", got)
	}
}

func TestLineageSettingsStatusHook_IgnoresOtherOperations(t *testing.T) {
	var n int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&n, 1)
	}))
	defer srv.Close()
	if err := lineageCall(t, newLineageHook(), context.Background(), srv.URL, "DeleteLineageWorkspaceSettings"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&n); got != 0 {
		t.Errorf("expected no HTTP calls, got %d", got)
	}
}

func TestLineageSettingsStatusHook_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"status":"PENDING"}`)
	}))
	defer srv.Close()
	h := newLineageHook()
	h.OverallTimeout = 5 * time.Second
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	err := lineageCall(t, h, ctx, srv.URL, "UpdateLineageWorkspaceSettings")
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
	if time.Since(start) > 2*time.Second {
		t.Errorf("hook did not stop promptly after cancel")
	}
}

func TestLineageSettingsStatusHook_EmptyStatusFailsFast(t *testing.T) {
	var n int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&n, 1)
		_, _ = fmt.Fprint(w, `{}`)
	}))
	defer srv.Close()
	h := newLineageHook()
	h.OverallTimeout = 5 * time.Second
	start := time.Now()
	err := lineageCall(t, h, context.Background(), srv.URL, "UpdateLineageWorkspaceSettings")
	if err == nil || !strings.Contains(err.Error(), "not saved") {
		t.Fatalf("expected error containing 'not saved', got %v", err)
	}
	if time.Since(start) > time.Second {
		t.Errorf("expected prompt failure, took %v", time.Since(start))
	}
	if got := atomic.LoadInt32(&n); got != 1 {
		t.Errorf("expected 1 GET, got %d", got)
	}
}
