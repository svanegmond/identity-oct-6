package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/svanegmond/agentic-eng-oct-6/internal/store"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrMissingAuthHeader  = errors.New("missing Authorization header")
	ErrInvalidAuthHeader  = errors.New("invalid Authorization header format; Bearer required")
)

type contextKey string

const claimsContextKey = contextKey("auth_claims")

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type Service struct {
	secretKey []byte
	store     store.DAO
}

func NewService(secretKey string, s store.DAO) *Service {
	return &Service{
		secretKey: []byte(secretKey),
		store:     s,
	}
}

func (s *Service) IssueToken(userID, username string, duration time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			Issuer:    "identity-go-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signed, nil
}

func (s *Service) VerifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secretKey, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// AuthenticateBearer extracts and verifies a Bearer token from the Authorization header value.
func (s *Service) AuthenticateBearer(authHeader string) (*Claims, error) {
	if authHeader == "" {
		return nil, ErrMissingAuthHeader
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, ErrInvalidAuthHeader
	}

	tokenString := strings.TrimSpace(parts[1])
	return s.VerifyToken(tokenString)
}

// AuthenticateRequest extracts and verifies a Bearer token from an incoming HTTP request.
func (s *Service) AuthenticateRequest(r *http.Request) (*Claims, error) {
	return s.AuthenticateBearer(r.Header.Get("Authorization"))
}

func (s *Service) Login(ctx context.Context, username, password string) (string, error) {
	cred, err := s.store.GetCredentialByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("failed to lookup credential: %w", err)
	}

	// Password check per SK-3 interview-mock latitude (exact match or hashed comparison)
	if cred.Password != password {
		return "", ErrInvalidCredentials
	}

	return s.IssueToken(cred.UserID, cred.Username, 24*time.Hour)
}

func ContextWithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

func ClaimsFromContext(ctx context.Context) *Claims {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}
