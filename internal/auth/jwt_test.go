package auth_test

import (
	"context"
	"errors"
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

func TestAuthService_AuthenticateBearer(t *testing.T) {
	svc := auth.NewService("mock-secret-key", &mockStore{})

	// Case 1: Missing Authorization header -> ErrMissingAuthHeader
	_, err := svc.AuthenticateBearer("")
	if !errors.Is(err, auth.ErrMissingAuthHeader) {
		t.Errorf("expected ErrMissingAuthHeader for empty header, got %v", err)
	}

	// Case 2: Non-bearer auth -> ErrInvalidAuthHeader
	_, err = svc.AuthenticateBearer("Basic dXNlcjpwYXNz")
	if !errors.Is(err, auth.ErrInvalidAuthHeader) {
		t.Errorf("expected ErrInvalidAuthHeader for basic auth, got %v", err)
	}

	// Case 3: Invalid token string -> ErrInvalidToken
	_, err = svc.AuthenticateBearer("Bearer invalid-token-string")
	if !errors.Is(err, auth.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken for invalid token, got %v", err)
	}

	// Case 4: Valid bearer token -> success
	validToken, err := svc.IssueToken("u1", "alice", time.Hour)
	if err != nil {
		t.Fatalf("failed to issue valid token: %v", err)
	}
	claims, err := svc.AuthenticateBearer("Bearer " + validToken)
	if err != nil {
		t.Fatalf("AuthenticateBearer failed for valid token: %v", err)
	}
	if claims.UserID != "u1" || claims.Username != "alice" {
		t.Errorf("claims mismatch: got %+v", claims)
	}

	// Case 5: AuthenticateRequest with valid HTTP request
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	reqClaims, err := svc.AuthenticateRequest(req)
	if err != nil {
		t.Fatalf("AuthenticateRequest failed: %v", err)
	}
	if reqClaims.UserID != "u1" || reqClaims.Username != "alice" {
		t.Errorf("claims mismatch from request: got %+v", reqClaims)
	}
}
