package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/svanegmond/agentic-eng-oct-6/internal/api"
	"github.com/svanegmond/agentic-eng-oct-6/internal/auth"
	"github.com/svanegmond/agentic-eng-oct-6/internal/store"
)

type memoryStore struct {
	profiles map[string]*store.UserProfile
	creds    map[string]*store.UserCredential
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		profiles: make(map[string]*store.UserProfile),
		creds:    make(map[string]*store.UserCredential),
	}
}

func (m *memoryStore) CreateProfile(ctx context.Context, p *store.UserProfile) error {
	m.profiles[p.ID] = p
	return nil
}

func (m *memoryStore) GetProfileByID(ctx context.Context, id string) (*store.UserProfile, error) {
	p, ok := m.profiles[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	return p, nil
}

func (m *memoryStore) SearchProfiles(ctx context.Context, q store.ProfileSearchQuery) ([]*store.UserProfile, error) {
	var results []*store.UserProfile
	for _, p := range m.profiles {
		matchName := q.Name == "" || strings.Contains(strings.ToLower(p.Name), strings.ToLower(q.Name))
		matchPhone := q.Phone == "" || strings.Contains(p.Phone, q.Phone)
		if matchName && matchPhone {
			results = append(results, p)
		}
	}
	return results, nil
}

func (m *memoryStore) CreateCredential(ctx context.Context, c *store.UserCredential) error {
	m.creds[c.Username] = c
	return nil
}

func (m *memoryStore) GetCredentialByUsername(ctx context.Context, username string) (*store.UserCredential, error) {
	c, ok := m.creds[username]
	if !ok {
		return nil, store.ErrNotFound
	}
	return c, nil
}

func (m *memoryStore) Close() error               { return nil }
func (m *memoryStore) Ping(ctx context.Context) error { return nil }

func setupAPITest(t *testing.T) (http.Handler, *auth.Service, *memoryStore) {
	memStore := newMemoryStore()
	authSvc := auth.NewService("api-test-secret-32-bytes-long!", memStore)
	handler := api.NewRouter(memStore, authSvc)
	return handler, authSvc, memStore
}

// TP-8: Unit/HTTP: Authenticated profile search and retrieve handlers (valid Bearer)
// return contracted profile shapes and error model per AC-8 / REST contract.
func TestAPI_ProfileSearchAndRetrieve_TP8(t *testing.T) {
	handler, authSvc, memStore := setupAPITest(t)
	ctx := context.Background()

	// Seed profile
	pID := uuid.NewString()
	p1 := &store.UserProfile{
		ID:        pID,
		Name:      "Diana Prince",
		Address:   "1 Themyscira Way",
		Phone:     "+15558889999",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		UpdatedAt: time.Now().UTC().Truncate(time.Second),
	}
	memStore.CreateProfile(ctx, p1)

	// Issue token
	token, err := authSvc.IssueToken("u1", "diana", time.Hour)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	// 1. Authenticated Retrieve: GET /profiles/{id} with valid Bearer
	req := httptest.NewRequest("GET", "/profiles/"+pID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /profiles/{id} status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	var retrievedProfile api.UserProfile
	if err := json.NewDecoder(rec.Body).Decode(&retrievedProfile); err != nil {
		t.Fatalf("failed to decode profile JSON: %v", err)
	}
	if retrievedProfile.Id != pID || retrievedProfile.Name != p1.Name || retrievedProfile.Phone != p1.Phone {
		t.Errorf("profile mismatch: got %+v, want %+v", retrievedProfile, p1)
	}

	// 2. Retrieve Not Found: GET /profiles/{nonexistent} -> 404
	req404 := httptest.NewRequest("GET", "/profiles/nonexistent-id", nil)
	req404.Header.Set("Authorization", "Bearer "+token)
	rec404 := httptest.NewRecorder()
	handler.ServeHTTP(rec404, req404)

	if rec404.Code != http.StatusNotFound {
		t.Errorf("GET /profiles/{nonexistent} status = %d, want 404", rec404.Code)
	}
	var errResp api.ErrorResponse
	if err := json.NewDecoder(rec404.Body).Decode(&errResp); err != nil || errResp.Error == "" {
		t.Errorf("expected ErrorResponse JSON for 404, got: %s", rec404.Body.String())
	}

	// 3. Authenticated Search: GET /profiles?name=Prince
	reqSearch := httptest.NewRequest("GET", "/profiles?name=Prince", nil)
	reqSearch.Header.Set("Authorization", "Bearer "+token)
	recSearch := httptest.NewRecorder()
	handler.ServeHTTP(recSearch, reqSearch)

	if recSearch.Code != http.StatusOK {
		t.Fatalf("GET /profiles?name=Prince status = %d, want 200, body: %s", recSearch.Code, recSearch.Body.String())
	}

	var searchResults []api.UserProfile
	if err := json.NewDecoder(recSearch.Body).Decode(&searchResults); err != nil {
		t.Fatalf("failed to decode search results: %v", err)
	}
	if len(searchResults) != 1 || searchResults[0].Id != pID {
		t.Errorf("expected 1 result with id %s, got %d results", pID, len(searchResults))
	}

	// 4. Refusal without Bearer: GET /profiles
	reqUnauth := httptest.NewRequest("GET", "/profiles", nil)
	recUnauth := httptest.NewRecorder()
	handler.ServeHTTP(recUnauth, reqUnauth)
	if recUnauth.Code != http.StatusUnauthorized {
		t.Errorf("GET /profiles without bearer status = %d, want 401", recUnauth.Code)
	}
}

// Test login & registration endpoints
func TestAPI_LoginAndRegister(t *testing.T) {
	handler, _, _ := setupAPITest(t)

	// 1. Register new user
	regPayload := map[string]string{
		"username": "clark",
		"password": "supersecretpassword",
		"name":     "Clark Kent",
		"address":  "344 Clinton St, Metropolis",
		"phone":    "+15550009999",
	}
	body, _ := json.Marshal(regPayload)
	reqReg := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
	reqReg.Header.Set("Content-Type", "application/json")
	recReg := httptest.NewRecorder()
	handler.ServeHTTP(recReg, reqReg)

	if recReg.Code != http.StatusCreated {
		t.Fatalf("POST /auth/register status = %d, want 201, body: %s", recReg.Code, recReg.Body.String())
	}

	var createdProfile api.UserProfile
	if err := json.NewDecoder(recReg.Body).Decode(&createdProfile); err != nil {
		t.Fatalf("failed to decode register response: %v", err)
	}
	if createdProfile.Name != "Clark Kent" {
		t.Errorf("registered profile mismatch: got %s", createdProfile.Name)
	}

	// 2. Login with credentials
	loginPayload := map[string]string{
		"username": "clark",
		"password": "supersecretpassword",
	}
	lbody, _ := json.Marshal(loginPayload)
	reqLogin := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(lbody))
	reqLogin.Header.Set("Content-Type", "application/json")
	recLogin := httptest.NewRecorder()
	handler.ServeHTTP(recLogin, reqLogin)

	if recLogin.Code != http.StatusOK {
		t.Fatalf("POST /auth/login status = %d, want 200, body: %s", recLogin.Code, recLogin.Body.String())
	}

	var loginResp api.LoginResponse
	if err := json.NewDecoder(recLogin.Body).Decode(&loginResp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	if loginResp.Token == "" {
		t.Errorf("expected non-empty token in login response")
	}

	// 3. Login with bad password -> 401
	badLoginPayload := map[string]string{
		"username": "clark",
		"password": "wrongpassword",
	}
	blbody, _ := json.Marshal(badLoginPayload)
	reqBadLogin := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(blbody))
	reqBadLogin.Header.Set("Content-Type", "application/json")
	recBadLogin := httptest.NewRecorder()
	handler.ServeHTTP(recBadLogin, reqBadLogin)

	if recBadLogin.Code != http.StatusUnauthorized {
		t.Errorf("POST /auth/login with bad password status = %d, want 401", recBadLogin.Code)
	}
}

// TestAPI_OpenAPIBearerAuthEnforcement verifies OpenAPI BearerAuth security is enforced via nethttp-middleware
// and confirms 401 responses use a stable unauthorized message without leaking jwt library parse details.
func TestAPI_OpenAPIBearerAuthEnforcement(t *testing.T) {
	handler, authSvc, _ := setupAPITest(t)

	// 1. Missing Authorization header on protected route -> 401 with stable message
	reqNoAuth := httptest.NewRequest("GET", "/profiles", nil)
	recNoAuth := httptest.NewRecorder()
	handler.ServeHTTP(recNoAuth, reqNoAuth)
	if recNoAuth.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing auth on protected route, got %d", recNoAuth.Code)
	}
	var errRespNoAuth api.ErrorResponse
	if err := json.NewDecoder(recNoAuth.Body).Decode(&errRespNoAuth); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errRespNoAuth.Error != "unauthorized" || errRespNoAuth.Message != "Unauthorized: missing or invalid bearer token" {
		t.Errorf("unexpected 401 response body for missing auth: %+v", errRespNoAuth)
	}

	// 2. Basic auth instead of Bearer on protected route -> 401 with stable message
	reqBasic := httptest.NewRequest("GET", "/profiles", nil)
	reqBasic.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	recBasic := httptest.NewRecorder()
	handler.ServeHTTP(recBasic, reqBasic)
	if recBasic.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for Basic auth on protected route, got %d", recBasic.Code)
	}
	var errRespBasic api.ErrorResponse
	if err := json.NewDecoder(recBasic.Body).Decode(&errRespBasic); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errRespBasic.Error != "unauthorized" || errRespBasic.Message != "Unauthorized: missing or invalid bearer token" {
		t.Errorf("unexpected 401 response body for basic auth: %+v", errRespBasic)
	}

	// 3. Corrupted Bearer token on protected route -> 401 with stable message (no jwt parse detail leak)
	reqCorrupt := httptest.NewRequest("GET", "/profiles", nil)
	reqCorrupt.Header.Set("Authorization", "Bearer invalid.jwt.token")
	recCorrupt := httptest.NewRecorder()
	handler.ServeHTTP(recCorrupt, reqCorrupt)
	if recCorrupt.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for corrupt token, got %d", recCorrupt.Code)
	}
	var errRespCorrupt api.ErrorResponse
	if err := json.NewDecoder(recCorrupt.Body).Decode(&errRespCorrupt); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errRespCorrupt.Error != "unauthorized" || errRespCorrupt.Message != "Unauthorized: missing or invalid bearer token" {
		t.Errorf("unexpected 401 response body for corrupt token: %+v", errRespCorrupt)
	}
	if strings.Contains(errRespCorrupt.Message, "token is malformed") || strings.Contains(errRespCorrupt.Message, "jwt") {
		t.Errorf("401 message must not echo jwt library details: %s", errRespCorrupt.Message)
	}

	// 4. Valid Bearer token on protected route -> 200 OK
	validToken, err := authSvc.IssueToken("u1", "testuser", time.Hour)
	if err != nil {
		t.Fatalf("failed to issue valid token: %v", err)
	}
	reqValid := httptest.NewRequest("GET", "/profiles", nil)
	reqValid.Header.Set("Authorization", "Bearer "+validToken)
	recValid := httptest.NewRecorder()
	handler.ServeHTTP(recValid, reqValid)
	if recValid.Code != http.StatusOK {
		t.Errorf("expected 200 for valid Bearer token on protected route, got %d", recValid.Code)
	}

	// 5. Unprotected route (/auth/login) does not enforce BearerAuth
	loginPayload := map[string]string{
		"username": "user",
		"password": "wrong",
	}
	body, _ := json.Marshal(loginPayload)
	reqPublic := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	reqPublic.Header.Set("Content-Type", "application/json")
	recPublic := httptest.NewRecorder()
	handler.ServeHTTP(recPublic, reqPublic)
	// Response should reach handler (which returns 401 for bad password, not 401 for missing Bearer)
	var errResp api.ErrorResponse
	_ = json.NewDecoder(recPublic.Body).Decode(&errResp)
	if errResp.Error != "unauthorized" || errResp.Message != "Invalid credentials" {
		t.Errorf("expected handler-level 401 invalid credentials, got %+v", errResp)
	}
}
