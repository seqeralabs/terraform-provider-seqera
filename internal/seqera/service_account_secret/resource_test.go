package service_account_secret

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk"
	"github.com/seqeralabs/terraform-provider-seqera/internal/sdk/models/shared"
)

type secret struct {
	ID        int64  `json:"id"`
	Hint      string `json:"hint"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	ExpiresAt string `json:"expiresAt"`
}

func s(id int64, status string) secret {
	return secret{ID: id, Hint: "****abcd", Status: status, CreatedAt: "2026-01-01T00:00:00Z", ExpiresAt: "2026-12-01T00:00:00Z"}
}

// newServer serves GET /orgs/1/service-accounts/2/secrets.
func newServer(t *testing.T, secrets []secret, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/orgs/1/service-accounts/2/secrets" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.String())
		}
		if status != http.StatusOK {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"message":"nope"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"secrets": secrets})
	}))
}

func TestFind(t *testing.T) {
	// During a rotation two secrets are live and the API lists the one that
	// expires first (the old one) first: find must match on ID, not position.
	rotation := []secret{s(10, "EXPIRING"), s(11, "ACTIVE"), s(9, "REVOKED")}

	cases := []struct {
		name     string
		secrets  []secret
		status   int
		lookup   int64
		wantID   int64
		wantNil  bool
		wantErr  bool
		wantLive bool
	}{
		{name: "new secret listed second", secrets: rotation, status: 200, lookup: 11, wantID: 11, wantLive: true},
		{name: "old secret listed first", secrets: rotation, status: 200, lookup: 10, wantID: 10, wantLive: true},
		{name: "revoked", secrets: rotation, status: 200, lookup: 9, wantID: 9},
		{name: "expired", secrets: []secret{s(8, "EXPIRED")}, status: 200, lookup: 8, wantID: 8},
		{name: "not listed", secrets: rotation, status: 200, lookup: 7, wantNil: true},
		{name: "service account gone", status: http.StatusNotFound, lookup: 7, wantNil: true},
		{name: "forbidden", status: http.StatusForbidden, lookup: 7, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newServer(t, tc.secrets, tc.status)
			defer srv.Close()
			client := sdk.New(sdk.WithServerURL(srv.URL), sdk.WithSecurity(shared.Security{BearerAuth: sdk.Pointer("test")}))

			got, err := find(context.Background(), client, 1, 2, tc.lookup)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if live(got) != tc.wantLive {
				t.Fatalf("live(%+v) = %v, want %v", got, !tc.wantLive, tc.wantLive)
			}
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected no match, got %+v", got)
				}
				return
			}
			if got == nil || got.ID != tc.wantID {
				t.Fatalf("expected id %d, got %+v", tc.wantID, got)
			}
		})
	}
}

func TestRefreshFormatsTimesAsRFC3339(t *testing.T) {
	srv := newServer(t, []secret{{ID: 5, Hint: "****wxyz", Status: "ACTIVE", CreatedAt: "2026-01-01T10:00:00.123+02:00", ExpiresAt: "2027-01-01T10:00:00Z"}}, 200)
	defer srv.Close()
	client := sdk.New(sdk.WithServerURL(srv.URL), sdk.WithSecurity(shared.Security{BearerAuth: sdk.Pointer("test")}))

	got, err := find(context.Background(), client, 1, 2, 5)
	if err != nil || got == nil {
		t.Fatalf("find: %v, %+v", err, got)
	}
	var data ResourceModel
	refresh(&data, got)

	if v := data.CreatedAt.ValueString(); v != "2026-01-01T08:00:00Z" {
		t.Errorf("created_at = %q", v)
	}
	if v := data.ExpiresAt.ValueString(); v != "2027-01-01T10:00:00Z" {
		t.Errorf("expires_at = %q", v)
	}
	if data.ID.ValueInt64() != 5 || data.Hint.ValueString() != "****wxyz" || data.Status.ValueString() != "ACTIVE" {
		t.Errorf("unexpected model %+v", data)
	}
}

func TestClientID(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		status  int
		want    string
		wantNil bool
		wantErr bool
	}{
		{name: "client exists", status: 200, body: `{"serviceAccount":{"id":2,"clientCredentials":{"clientId":"sa_abc","status":"ACTIVE"}}}`, want: "sa_abc"},
		{name: "no client yet", status: 200, body: `{"serviceAccount":{"id":2,"clientCredentials":{"status":"NONE"}}}`, wantNil: true},
		{name: "credentials disabled", status: 200, body: `{"serviceAccount":{"id":2}}`, wantNil: true},
		{name: "not found", status: 404, body: `{"message":"nope"}`, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/orgs/1/service-accounts/2" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.String())
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()
			client := sdk.New(sdk.WithServerURL(srv.URL), sdk.WithSecurity(shared.Security{BearerAuth: sdk.Pointer("test")}))

			got, err := clientID(context.Background(), client, 1, 2)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %q", *got)
				}
				return
			}
			if got == nil || *got != tc.want {
				t.Fatalf("expected %q, got %v", tc.want, got)
			}
		})
	}
}
