package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/svanegmond/agentic-eng-oct-6/internal/auth"
	"github.com/svanegmond/agentic-eng-oct-6/internal/store"
)

type mockStore struct {
	creds map[string]*store.UserCredential
}

func (m *mockStore) CreateProfile(ctx context.Context, p *store.UserProfile) error { return nil }
func (m *mockStore) GetProfileByID(ctx context.Context, id string) (*store.UserProfile, error) {
	return nil, store.ErrNotFound
}
func (m *mockStore) SearchProfiles(ctx context.Context, q store.ProfileSearchQuery) ([]*store.UserProfile, error) {
	return nil, nil
}
func (m *mockStore) CreateCredential(ctx context.Context, c *store.UserCredential) error {
	m.creds[c.Username] = c
	return nil
}
func (m *mockStore) GetCredentialByUsername(ctx context.Context, username string) (*store.UserCredential, error) {
	c, ok := m.creds[username]
	if !ok {
		return nil, store.ErrNotFound
	}
	return c, nil
}
func (m *mockStore) Close() error               { return nil }
func (m *mockStore) Ping(ctx context.Context) error { return nil }

func TestAuthService_TokenIssueAndVerify(t *testing.T) {
	secretKey := "test-secret-key-32-bytes-long!!"
	svc := auth.NewService(secretKey, &mockStore{})

	// 1. Issue valid token
	userID := uuid.NewString()
	token, err := svc.IssueToken(userID, "alice", time.Hour)
	if err != nil {
		t.Fatalf("IssueToken failed: %v", err)
	}
	if token == "" {
		t.Fatalf("expected non-empty token")
	}

	// 2. Verify valid token
	claims, err := svc.VerifyToken(token)
	if err != nil {
		t.Fatalf("VerifyToken failed: %v", err)
	}
	if claims.Username != "alice" || claims.UserID != userID {
		t.Errorf("claims mismatch: got %+v", claims)
	}

	// 3. Reject invalid signature
	svcDiffSecret := auth.NewService("different-secret-key-long-enough", &mockStore{})
	_, err = svcDiffSecret.VerifyToken(token)
	if err == nil {
		t.Errorf("expected error verifying token with different secret, got nil")
	}

	// 4. Reject expired token
	expiredToken, err := svc.IssueToken(userID, "alice", -time.Minute)
	if err != nil {
		t.Fatalf("failed to issue expired token: %v", err)
	}
	_, err = svc.VerifyToken(expiredToken)
	if err == nil {
		t.Errorf("expected error verifying expired token, got nil")
	}
}

func TestAuthService_CredentialCheckAndLogin(t *testing.T) {
	ctx := context.Background()
	mock := &mockStore{creds: make(map[string]*store.UserCredential)}
	mock.CreateCredential(ctx, &store.UserCredential{
		ID:       uuid.NewString(),
		UserID:   "user-123",
		Username: "alice",
		Method:   "password",
		Password: "correct-password",
	})

	svc := auth.NewService("mock-secret-key", mock)

	// Valid login
	token, err := svc.Login(ctx, "alice", "correct-password")
	if err != nil {
		t.Fatalf("Login with correct credentials failed: %v", err)
	}
	claims, err := svc.VerifyToken(token)
	if err != nil || claims.Username != "alice" {
		t.Fatalf("VerifyToken on issued login token failed: %v", err)
	}

	// Invalid password
	_, err = svc.Login(ctx, "alice", "wrong-password")
	if err == nil {
		t.Errorf("expected error for wrong password, got nil")
	}

	// Nonexistent user
	_, err = svc.Login(ctx, "nonexistent", "correct-password")
	if err == nil {
		t.Errorf("expected error for nonexistent user, got nil")
	}
}

func TestAuthMiddleware_Protection(t *testing.T) {
	svc := auth.NewService("mock-secret-key", &mockStore{})

	// Protected handler returns 200 OK
	protectedHandler := svc.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := auth.ClaimsFromContext(r.Context())
		if claims == nil {
			t.Errorf("expected claims in context")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))

	// Case 1: Missing Authorization header -> 401
	req1 := httptest.NewRequest("GET", "/protected", nil)
	rec1 := httptest.NewRecorder()
	protectedHandler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing auth header, got %d", rec1.Code)
	}
	var errResp map[string]interface{}
	if err := json.NewDecoder(rec1.Body).Decode(&errResp); err != nil || errResp["error"] == nil {
		t.Errorf("expected error response JSON, got %v", rec1.Body.String())
	}

	// Case 2: Invalid bearer format -> 401
	req2 := httptest.NewRequest("GET", "/protected", nil)
	req2.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rec2 := httptest.NewRecorder()
	protectedHandler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for non-bearer auth, got %d", rec2.Code)
	}

	// Case 3: Invalid token string -> 401
	req3 := httptest.NewRequest("GET", "/protected", nil)
	req3.Header.Set("Authorization", "Bearer invalid-token-string")
	rec3 := httptest.NewRecorder()
	protectedHandler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid bearer token, got %d", rec3.Code)
	}

	// Case 4: Valid bearer token -> 200
	validToken, err := svc.IssueToken("u1", "alice", time.Hour)
	if err != nil {
		t.Fatalf("failed to issue valid token: %v", err)
	}
	req4 := httptest.NewRequest("GET", "/protected", nil)
	req4.Header.Set("Authorization", "Bearer "+validToken)
	rec4 := httptest.NewRecorder()
	protectedHandler.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusOK {
		t.Errorf("expected 200 for valid bearer token, got %d", rec4.Code)
	}
}
