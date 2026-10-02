package agent_data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/shared"
)

type agent struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status,omitempty"`
}

// newServer serves GET /agents?workspaceId=5&search=..., mimicking the API's
// substring search, and honours max/offset.
func newServer(t *testing.T, all []agent, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.URL.Path != "/agents" || q.Get("workspaceId") != "5" {
			t.Errorf("unexpected request %s", r.URL.String())
		}
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		matched := []agent{}
		for _, a := range all {
			if strings.Contains(a.Name, q.Get("search")) {
				matched = append(matched, a)
			}
		}
		offset, _ := strconv.Atoi(q.Get("offset"))
		max, _ := strconv.Atoi(q.Get("max"))
		page := []agent{}
		if offset < len(matched) {
			page = matched[offset:min(offset+max, len(matched))]
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"agents": page, "totalSize": len(matched)})
	}))
}

func decoys(n int) []agent {
	out := make([]agent, n)
	for i := range out {
		out[i] = agent{ID: fmt.Sprintf("d%03d", i), Name: fmt.Sprintf("ci-%03d", i), Status: "active"}
	}
	return out
}

func TestFind(t *testing.T) {
	cases := []struct {
		name    string
		agents  []agent
		status  int
		lookup  string
		wantID  string
		wantNil bool
		wantErr bool
	}{
		{name: "first page", agents: []agent{{ID: "a1", Name: "ci", Status: "active"}}, status: 200, lookup: "ci", wantID: "a1"},
		{name: "second page", agents: append(decoys(100), agent{ID: "a2", Name: "ci", Status: "inactive"}), status: 200, lookup: "ci", wantID: "a2"},
		{name: "substring decoy", agents: []agent{{ID: "n1", Name: "ci-nightly"}, {ID: "a3", Name: "ci"}}, status: 200, lookup: "ci", wantID: "a3"},
		{name: "not found", agents: []agent{{ID: "n1", Name: "ci-nightly"}}, status: 200, lookup: "ci", wantNil: true},
		{name: "forbidden", status: http.StatusForbidden, lookup: "ci", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newServer(t, tc.agents, tc.status)
			defer srv.Close()
			client := sdk.New(sdk.WithServerURL(srv.URL), sdk.WithSecurity(shared.Security{BearerAuth: "test"}))

			got, err := find(context.Background(), client, 5, tc.lookup)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected no match, got %+v", got)
				}
				return
			}
			if got == nil || got.ID == nil || *got.ID != tc.wantID {
				t.Fatalf("expected id %s, got %+v", tc.wantID, got)
			}
		})
	}
}
