package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/svanegmond/agentic-eng-oct-6/internal/store"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPostgresDAO_StoreRetrieveSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgres testcontainers integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("identity_test"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	defer func() {
		if err := pgContainer.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	dao, err := store.Open(ctx, store.DBConfig{
		Driver: "postgres",
		DSN:    connStr,
	})
	if err != nil {
		t.Fatalf("failed to open postgres dao: %v", err)
	}
	defer dao.Close()

	if err := dao.Ping(ctx); err != nil {
		t.Fatalf("failed to ping postgres dao: %v", err)
	}

	// 1. Create Profile
	profileID := uuid.NewString()
	p1 := &store.UserProfile{
		ID:        profileID,
		Name:      "Carol Danvers",
		Address:   "789 Star St, Marvel",
		Phone:     "+15554321098",
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
		Name:      "Danvers Smith",
		Address:   "101 Moon Rd, Marvel",
		Phone:     "+15557654321",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		UpdatedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := dao.CreateProfile(ctx, p2); err != nil {
		t.Fatalf("CreateProfile p2 failed: %v", err)
	}

	// 4. Search Profiles by Name ("danvers" case-insensitive)
	results, err := dao.SearchProfiles(ctx, store.ProfileSearchQuery{Name: "danvers"})
	if err != nil {
		t.Fatalf("SearchProfiles by name failed: %v", err)
	}
	if len(results) < 2 {
		t.Errorf("SearchProfiles by name expected >= 2 results, got %d", len(results))
	}

	// 5. Search Profiles by Phone ("7654")
	phoneResults, err := dao.SearchProfiles(ctx, store.ProfileSearchQuery{Phone: "7654"})
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
		Username:  "carol",
		Method:    "password",
		Password:  "hashed_secret_carol",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		UpdatedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := dao.CreateCredential(ctx, cred); err != nil {
		t.Fatalf("CreateCredential failed: %v", err)
	}

	// 7. Retrieve Credential by Username
	gotCred, err := dao.GetCredentialByUsername(ctx, "carol")
	if err != nil {
		t.Fatalf("GetCredentialByUsername failed: %v", err)
	}
	if gotCred.ID != credID || gotCred.UserID != profileID || gotCred.Password != cred.Password {
		t.Errorf("GetCredentialByUsername mismatch: got %+v, want %+v", gotCred, cred)
	}

	// 8. Verify not found error
	_, err = dao.GetCredentialByUsername(ctx, "nonexistent")
	if err == nil {
		t.Errorf("expected error for nonexistent credential, got nil")
	}
}
