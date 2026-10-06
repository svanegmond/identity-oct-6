package store

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/svanegmond/agentic-eng-oct-6/internal/store/sqlite/sqlc_sqlite"
	_ "modernc.org/sqlite"
)

//go:embed sqlite/migrations/*.sql
var sqliteMigrationsFS embed.FS

type sqliteDAO struct {
	db      *sql.DB
	queries *sqlc_sqlite.Queries
}

func openSQLite(ctx context.Context, dsn string) (DAO, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Apply migrations
	goose.SetBaseFS(sqliteMigrationsFS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set goose dialect for sqlite: %w", err)
	}
	if err := goose.Up(db, "sqlite/migrations"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run sqlite migrations: %w", err)
	}

	return &sqliteDAO{
		db:      db,
		queries: sqlc_sqlite.New(db),
	}, nil
}

func (s *sqliteDAO) Close() error {
	return s.db.Close()
}

func (s *sqliteDAO) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *sqliteDAO) CreateProfile(ctx context.Context, p *UserProfile) error {
	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = now
	}

	err := s.queries.CreateProfile(ctx, sqlc_sqlite.CreateProfileParams{
		ID:        p.ID,
		Name:      p.Name,
		Address:   p.Address,
		Phone:     p.Phone,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}
	return nil
}

func (s *sqliteDAO) GetProfileByID(ctx context.Context, id string) (*UserProfile, error) {
	row, err := s.queries.GetProfileByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get profile by id: %w", err)
	}
	return &UserProfile{
		ID:        row.ID,
		Name:      row.Name,
		Address:   row.Address,
		Phone:     row.Phone,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (s *sqliteDAO) SearchProfiles(ctx context.Context, q ProfileSearchQuery) ([]*UserProfile, error) {
	rows, err := s.queries.SearchProfiles(ctx, sqlc_sqlite.SearchProfilesParams{
		Name:  q.Name,
		Phone: q.Phone,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search profiles: %w", err)
	}
	profiles := make([]*UserProfile, len(rows))
	for i, r := range rows {
		profiles[i] = &UserProfile{
			ID:        r.ID,
			Name:      r.Name,
			Address:   r.Address,
			Phone:     r.Phone,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		}
	}
	return profiles, nil
}

func (s *sqliteDAO) CreateCredential(ctx context.Context, c *UserCredential) error {
	now := time.Now().UTC()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = now
	}

	err := s.queries.CreateCredential(ctx, sqlc_sqlite.CreateCredentialParams{
		ID:        c.ID,
		UserID:    c.UserID,
		Username:  c.Username,
		Method:    c.Method,
		Password:  c.Password,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}
	return nil
}

func (s *sqliteDAO) GetCredentialByUsername(ctx context.Context, username string) (*UserCredential, error) {
	row, err := s.queries.GetCredentialByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}
	return &UserCredential{
		ID:        row.ID,
		UserID:    row.UserID,
		Username:  row.Username,
		Method:    row.Method,
		Password:  row.Password,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}
