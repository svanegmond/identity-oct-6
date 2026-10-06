package store

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type UserProfile struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserCredential struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Method    string    `json:"method"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProfileSearchQuery struct {
	Name  string
	Phone string
}

type DBConfig struct {
	Driver string // "sqlite" or "postgres"
	DSN    string // connection string or file path
}

// DAO defines data access for user profiles and credentials.
type DAO interface {
	// Profile operations
	CreateProfile(ctx context.Context, p *UserProfile) error
	GetProfileByID(ctx context.Context, id string) (*UserProfile, error)
	SearchProfiles(ctx context.Context, q ProfileSearchQuery) ([]*UserProfile, error)

	// Credential operations
	CreateCredential(ctx context.Context, c *UserCredential) error
	GetCredentialByUsername(ctx context.Context, username string) (*UserCredential, error)

	// Lifecycle
	Close() error
	Ping(ctx context.Context) error
}

func Open(ctx context.Context, cfg DBConfig) (DAO, error) {
	switch cfg.Driver {
	case "sqlite":
		return openSQLite(ctx, cfg.DSN)
	case "postgres":
		return openPostgres(ctx, cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported database driver: %q", cfg.Driver)
	}
}

// openPostgres is implemented in postgres_adapter.go or postgres/store.go
var openPostgres = func(ctx context.Context, dsn string) (DAO, error) {
	return nil, errors.New("postgres adapter not yet initialized")
}
