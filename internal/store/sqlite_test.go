package store_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/svanegmond/agentic-eng-oct-6/internal/store"
)

func TestSQLiteDAO_StoreRetrieveSearch(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_identity.db")

	dao, err := store.Open(ctx, store.DBConfig{
		Driver: "sqlite",
		DSN:    dbPath,
	})
	if err != nil {
		t.Fatalf("failed to open sqlite dao: %v", err)
	}
	defer dao.Close()

	if err := dao.Ping(ctx); err != nil {
		t.Fatalf("failed to ping sqlite dao: %v", err)
	}

	// 1. Create Profile
	profileID := uuid.NewString()
	p1 := &store.UserProfile{
		ID:        profileID,
		Name:      "Alice Smith",
		Address:   "123 Main St, Springfield",
		Phone:     "+15551234567",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		UpdatedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := dao.CreateProfile(ctx, p1); err != nil {
		t.Fatalf("CreateProfile failed: %v", err)
	}

	// 2. Retrieve Profile by ID
	gotP1, err := dao.GetProfileByID(ctx, profileID)
	if err != nil {
		t.Fatalf("GetProfileByID failed: %v", err)
	}
	if gotP1.Name != p1.Name || gotP1.Phone != p1.Phone || gotP1.Address != p1.Address {
		t.Errorf("GetProfileByID mismatch: got %+v, want %+v", gotP1, p1)
	}

	// 3. Store another profile for search
	p2 := &store.UserProfile{
		ID:        uuid.NewString(),
		Name:      "Bob Smith",
		Address:   "456 Oak Ave, Springfield",
		Phone:     "+15559876543",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		UpdatedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := dao.CreateProfile(ctx, p2); err != nil {
		t.Fatalf("CreateProfile p2 failed: %v", err)
	}

	// 4. Search Profiles by Name ("Smith")
	results, err := dao.SearchProfiles(ctx, store.ProfileSearchQuery{Name: "Smith"})
	if err != nil {
		t.Fatalf("SearchProfiles by name failed: %v", err)
	}
	if len(results) < 2 {
		t.Errorf("SearchProfiles by name expected >= 2 results, got %d", len(results))
	}

	// 5. Search Profiles by Phone ("9876")
	phoneResults, err := dao.SearchProfiles(ctx, store.ProfileSearchQuery{Phone: "9876"})
	if err != nil {
		t.Fatalf("SearchProfiles by phone failed: %v", err)
	}
	if len(phoneResults) != 1 || phoneResults[0].ID != p2.ID {
		t.Errorf("SearchProfiles by phone expected 1 result for p2, got %d", len(phoneResults))
	}

	// 6. Create Credential
	credID := uuid.NewString()
	cred := &store.UserCredential{
		ID:        credID,
		UserID:    profileID,
		Username:  "alice",
		Method:    "password",
		Password:  "hashed_secret_alice",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		UpdatedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := dao.CreateCredential(ctx, cred); err != nil {
		t.Fatalf("CreateCredential failed: %v", err)
	}

	// 7. Retrieve Credential by Username
	gotCred, err := dao.GetCredentialByUsername(ctx, "alice")
	if err != nil {
		t.Fatalf("GetCredentialByUsername failed: %v", err)
	}
	if gotCred.ID != credID || gotCred.UserID != profileID || gotCred.Password != cred.Password {
		t.Errorf("GetCredentialByUsername mismatch: got %+v, want %+v", gotCred, cred)
	}

	// Verify not found for missing credential
	_, err = dao.GetCredentialByUsername(ctx, "nonexistent")
	if err == nil {
		t.Errorf("expected error for nonexistent credential, got nil")
	}

	// Verify file is created and durable
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("sqlite database file was not created on disk: %s", dbPath)
	}
}
