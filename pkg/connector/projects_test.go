package connector

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/atlas-sdk/v20250312006/admin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	testProjectID = "P"
	testUserID    = "U"
)

type mockResponse struct {
	status int
	body   any
}

type recordedCall struct {
	method string
	path   string
	body   map[string]any
}

type mockAtlas struct {
	mu        sync.Mutex
	calls     []recordedCall
	responses map[string]mockResponse
}

func (m *mockAtlas) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	call := recordedCall{method: r.Method, path: r.URL.Path}
	if data, _ := io.ReadAll(r.Body); len(data) > 0 {
		_ = json.Unmarshal(data, &call.body)
	}

	m.mu.Lock()
	m.calls = append(m.calls, call)
	m.mu.Unlock()

	resp, ok := m.responses[r.Method+" "+r.URL.Path]
	if !ok {
		resp = mockResponse{status: http.StatusNotImplemented, body: apiError(http.StatusNotImplemented, "UNEXPECTED_CALL")}
	}

	w.Header().Set("Content-Type", "application/vnd.atlas.2025-03-12+json")
	w.WriteHeader(resp.status)
	if resp.body != nil {
		_ = json.NewEncoder(w).Encode(resp.body)
	}
}

func (m *mockAtlas) mutations() []recordedCall {
	var rv []recordedCall
	for _, c := range m.calls {
		if c.method != http.MethodGet {
			rv = append(rv, c)
		}
	}
	return rv
}

func apiError(code int, errorCode string) map[string]any {
	return map[string]any{
		"error":     code,
		"errorCode": errorCode,
		"detail":    errorCode,
		"reason":    http.StatusText(code),
	}
}

func projectUser(roles ...string) map[string]any {
	return map[string]any{
		"id":                  testUserID,
		"username":            "user@example.com",
		"orgMembershipStatus": "ACTIVE",
		"roles":               roles,
	}
}

func newTestProjectBuilder(t *testing.T, responses map[string]mockResponse) (*projectBuilder, *mockAtlas) {
	t.Helper()

	mock := &mockAtlas{responses: responses}
	srv := httptest.NewServer(mock)
	t.Cleanup(srv.Close)

	client, err := admin.NewClient(admin.UseBaseURL(srv.URL))
	require.NoError(t, err)

	return newProjectBuilder(client), mock
}

func testPrincipal(t *testing.T) *v2.Resource {
	t.Helper()

	principal, err := rs.NewUserResource(
		"user@example.com",
		userResourceType,
		testUserID,
		[]rs.UserTraitOption{rs.WithUserLogin("user@example.com")},
	)
	require.NoError(t, err)

	return principal
}

func testEntitlement(slug string) *v2.Entitlement {
	return &v2.Entitlement{
		Id:   "project:" + testProjectID + ":" + slug,
		Slug: slug,
		Resource: &v2.Resource{
			Id: &v2.ResourceId{ResourceType: projectResourceType.Id, Resource: testProjectID},
		},
	}
}

const (
	userPath       = "/api/atlas/v2/groups/P/users/U"
	usersPath      = "/api/atlas/v2/groups/P/users"
	addRolePath    = "/api/atlas/v2/groups/P/users/U:addRole"
	removeRolePath = "/api/atlas/v2/groups/P/users/U:removeRole"
)

func TestProjectGrant(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		responses     map[string]mockResponse
		slug          string
		wantMutations []recordedCall
		wantExists    bool
		wantCode      codes.Code
	}{
		{
			name: "user not in project is added",
			responses: map[string]mockResponse{
				"GET " + userPath:   {http.StatusNotFound, apiError(http.StatusNotFound, "USER_NOT_IN_GROUP")},
				"POST " + usersPath: {http.StatusOK, projectUser("GROUP_OWNER")},
			},
			slug: ownerEntitlement,
			wantMutations: []recordedCall{{
				method: http.MethodPost,
				path:   usersPath,
				body:   map[string]any{"username": "user@example.com", "roles": []any{"GROUP_OWNER"}},
			}},
		},
		{
			name: "existing member gets an additional role",
			responses: map[string]mockResponse{
				"GET " + userPath:     {http.StatusOK, projectUser("GROUP_DATA_ACCESS_READ_WRITE")},
				"POST " + addRolePath: {http.StatusOK, projectUser("GROUP_DATA_ACCESS_READ_WRITE", "GROUP_OWNER")},
			},
			slug: ownerEntitlement,
			wantMutations: []recordedCall{{
				method: http.MethodPost,
				path:   addRolePath,
				body:   map[string]any{"groupRole": "GROUP_OWNER"},
			}},
		},
		{
			name: "role already granted",
			responses: map[string]mockResponse{
				"GET " + userPath: {http.StatusOK, projectUser("GROUP_OWNER")},
			},
			slug:       ownerEntitlement,
			wantExists: true,
		},
		{
			name: "user joined concurrently falls back to add role",
			responses: map[string]mockResponse{
				"GET " + userPath:     {http.StatusNotFound, apiError(http.StatusNotFound, "USER_NOT_IN_GROUP")},
				"POST " + usersPath:   {http.StatusBadRequest, apiError(http.StatusBadRequest, "USER_ALREADY_IN_GROUP")},
				"POST " + addRolePath: {http.StatusOK, projectUser("GROUP_OWNER")},
			},
			slug: ownerEntitlement,
			wantMutations: []recordedCall{
				{
					method: http.MethodPost,
					path:   usersPath,
					body:   map[string]any{"username": "user@example.com", "roles": []any{"GROUP_OWNER"}},
				},
				{
					method: http.MethodPost,
					path:   addRolePath,
					body:   map[string]any{"groupRole": "GROUP_OWNER"},
				},
			},
		},
		{
			name: "lookup forbidden returns error without mutation",
			responses: map[string]mockResponse{
				"GET " + userPath: {http.StatusForbidden, apiError(http.StatusForbidden, "USER_UNAUTHORIZED")},
			},
			slug:     ownerEntitlement,
			wantCode: codes.PermissionDenied,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			builder, mock := newTestProjectBuilder(t, tc.responses)

			annos, err := builder.Grant(ctx, testPrincipal(t), testEntitlement(tc.slug))

			if tc.wantCode != codes.OK {
				require.Error(t, err)
				require.Equal(t, tc.wantCode, status.Code(err))
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.wantExists, annos.Contains(&v2.GrantAlreadyExists{}))
			require.Equal(t, tc.wantMutations, mock.mutations())
		})
	}
}

func TestProjectRevoke(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		responses     map[string]mockResponse
		slug          string
		wantMutations []recordedCall
		wantRevoked   bool
		wantCode      codes.Code
	}{
		{
			name: "one of several roles is removed",
			responses: map[string]mockResponse{
				"GET " + userPath:        {http.StatusOK, projectUser("GROUP_DATA_ACCESS_READ_WRITE", "GROUP_OWNER")},
				"POST " + removeRolePath: {http.StatusOK, projectUser("GROUP_DATA_ACCESS_READ_WRITE")},
			},
			slug: ownerEntitlement,
			wantMutations: []recordedCall{{
				method: http.MethodPost,
				path:   removeRolePath,
				body:   map[string]any{"groupRole": "GROUP_OWNER"},
			}},
		},
		{
			name: "last role removes user from project",
			responses: map[string]mockResponse{
				"GET " + userPath:    {http.StatusOK, projectUser("GROUP_OWNER")},
				"DELETE " + userPath: {http.StatusNoContent, nil},
			},
			slug:          ownerEntitlement,
			wantMutations: []recordedCall{{method: http.MethodDelete, path: userPath}},
		},
		{
			name: "role not held",
			responses: map[string]mockResponse{
				"GET " + userPath: {http.StatusOK, projectUser("GROUP_DATA_ACCESS_READ_WRITE")},
			},
			slug:        ownerEntitlement,
			wantRevoked: true,
		},
		{
			name: "user not in project",
			responses: map[string]mockResponse{
				"GET " + userPath: {http.StatusNotFound, apiError(http.StatusNotFound, "USER_NOT_IN_GROUP")},
			},
			slug:        ownerEntitlement,
			wantRevoked: true,
		},
		{
			name: "rate limited lookup returns error",
			responses: map[string]mockResponse{
				"GET " + userPath: {http.StatusTooManyRequests, apiError(http.StatusTooManyRequests, "RATE_LIMITED")},
			},
			slug:     ownerEntitlement,
			wantCode: codes.Unavailable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			builder, mock := newTestProjectBuilder(t, tc.responses)

			grant := &v2.Grant{
				Entitlement: testEntitlement(tc.slug),
				Principal:   testPrincipal(t),
			}

			annos, err := builder.Revoke(ctx, grant)

			if tc.wantCode != codes.OK {
				require.Error(t, err)
				require.Equal(t, tc.wantCode, status.Code(err))
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.wantRevoked, annos.Contains(&v2.GrantAlreadyRevoked{}))
			require.Equal(t, tc.wantMutations, mock.mutations())
		})
	}
}
