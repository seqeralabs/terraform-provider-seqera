package service_account_data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/shared"
)

type sa struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	MemberID int64  `json:"memberId"`
}

// newServer serves GET /orgs/7/service-accounts from all, honouring max/offset.
func newServer(t *testing.T, all []sa, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orgs/7/service-accounts" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		max, _ := strconv.Atoi(r.URL.Query().Get("max"))
		end := min(offset+max, len(all))
		page := []sa{}
		if offset < len(all) {
			page = all[offset:end]
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"serviceAccounts": page, "totalSize": len(all)})
	}))
}

func filler(n int) []sa {
	out := make([]sa, n)
	for i := range out {
		out[i] = sa{ID: int64(1000 + i), Name: fmt.Sprintf("filler-%03d", i), MemberID: int64(2000 + i)}
	}
	return out
}

func TestFind(t *testing.T) {
	cases := []struct {
		name     string
		accounts []sa
		status   int
		lookup   string
		wantID   int64
		wantNil  bool
		wantErr  bool
	}{
		{name: "first page", accounts: append(filler(3), sa{ID: 42, Name: "ci", MemberID: 99}), status: 200, lookup: "ci", wantID: 42},
		{name: "second page", accounts: append(filler(100), sa{ID: 43, Name: "ci", MemberID: 98}), status: 200, lookup: "ci", wantID: 43},
		{name: "exact match only", accounts: []sa{{ID: 1, Name: "ci-nightly"}, {ID: 2, Name: "ci"}}, status: 200, lookup: "ci", wantID: 2},
		{name: "not found", accounts: filler(5), status: 200, lookup: "ci", wantNil: true},
		{name: "forbidden", status: http.StatusForbidden, lookup: "ci", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newServer(t, tc.accounts, tc.status)
			defer srv.Close()
			client := sdk.New(sdk.WithServerURL(srv.URL), sdk.WithSecurity(shared.Security{BearerAuth: "test"}))

			got, err := find(context.Background(), client, 7, tc.lookup)

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
				t.Fatalf("expected id %d, got %+v", tc.wantID, got)
			}
		})
	}
}
