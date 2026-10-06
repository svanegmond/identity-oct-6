package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/svanegmond/agentic-eng-oct-6/internal/auth"
	"github.com/svanegmond/agentic-eng-oct-6/internal/idp"
	"github.com/svanegmond/agentic-eng-oct-6/internal/store"
)

type Server struct {
	store   store.DAO
	authSvc *auth.Service
	idpConn idp.Connector
}

func NewServer(s store.DAO, a *auth.Service, i idp.Connector) *Server {
	return &Server{
		store:   s,
		authSvc: a,
		idpConn: i,
	}
}

// Ensure Server implements StrictServerInterface
var _ StrictServerInterface = (*Server)(nil)

func (s *Server) Login(ctx context.Context, request LoginRequestObject) (LoginResponseObject, error) {
	if request.Body == nil {
		return Login401JSONResponse{Error: "invalid_request", Message: "Request body required"}, nil
	}

	token, err := s.authSvc.Login(ctx, request.Body.Username, request.Body.Password)
	if err != nil {
		return Login401JSONResponse{Error: "unauthorized", Message: "Invalid credentials"}, nil
	}

	return Login200JSONResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: 86400,
	}, nil
}

func (s *Server) Register(ctx context.Context, request RegisterRequestObject) (RegisterResponseObject, error) {
	if request.Body == nil {
		return Register400JSONResponse{Error: "bad_request", Message: "Request body required"}, nil
	}

	req := request.Body
	if req.Username == "" || req.Password == "" || req.Name == "" {
		return Register400JSONResponse{Error: "bad_request", Message: "username, password, and name are required"}, nil
	}

	// Check if username already exists
	_, err := s.store.GetCredentialByUsername(ctx, req.Username)
	if err == nil {
		return Register409JSONResponse{Error: "conflict", Message: "Username already exists"}, nil
	}

	profileID := uuid.NewString()
	now := time.Now().UTC()
	p := &store.UserProfile{
		ID:        profileID,
		Name:      req.Name,
		Address:   req.Address,
		Phone:     req.Phone,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.CreateProfile(ctx, p); err != nil {
		return Register400JSONResponse{Error: "internal_error", Message: fmt.Sprintf("failed to create profile: %v", err)}, nil
	}

	c := &store.UserCredential{
		ID:        uuid.NewString(),
		UserID:    profileID,
		Username:  req.Username,
		Method:    "password",
		Password:  req.Password,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.CreateCredential(ctx, c); err != nil {
		return Register400JSONResponse{Error: "internal_error", Message: fmt.Sprintf("failed to create credential: %v", err)}, nil
	}

	return Register201JSONResponse{
		Id:        p.ID,
		Name:      p.Name,
		Address:   p.Address,
		Phone:     p.Phone,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}, nil
}

func (s *Server) GetProfileByID(ctx context.Context, request GetProfileByIDRequestObject) (GetProfileByIDResponseObject, error) {
	p, err := s.store.GetProfileByID(ctx, request.Id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return GetProfileByID404JSONResponse{
				Error:   "not_found",
				Message: "Profile not found",
			}, nil
		}
		return nil, err
	}

	return GetProfileByID200JSONResponse{
		Id:        p.ID,
		Name:      p.Name,
		Address:   p.Address,
		Phone:     p.Phone,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}, nil
}

func (s *Server) SearchProfiles(ctx context.Context, request SearchProfilesRequestObject) (SearchProfilesResponseObject, error) {
	var nameQuery, phoneQuery string
	if request.Params.Name != nil {
		nameQuery = *request.Params.Name
	}
	if request.Params.Phone != nil {
		phoneQuery = *request.Params.Phone
	}

	profiles, err := s.store.SearchProfiles(ctx, store.ProfileSearchQuery{
		Name:  nameQuery,
		Phone: phoneQuery,
	})
	if err != nil {
		return nil, err
	}

	response := make([]UserProfile, len(profiles))
	for i, p := range profiles {
		response[i] = UserProfile{
			Id:        p.ID,
			Name:      p.Name,
			Address:   p.Address,
			Phone:     p.Phone,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		}
	}

	return SearchProfiles200JSONResponse(response), nil
}

func (s *Server) EnrichProfile(ctx context.Context, request EnrichProfileRequestObject) (EnrichProfileResponseObject, error) {
	if request.Body == nil {
		return EnrichProfile502JSONResponse{Error: "bad_request", Message: "Request body required"}, nil
	}

	pii, err := s.idpConn.FetchIdentity(ctx, request.Body.Name, request.Body.Phone)
	if err != nil {
		return EnrichProfile502JSONResponse{
			Error:   "upstream_idp_error",
			Message: fmt.Sprintf("Failed to fetch PII from IdP: %v", err),
		}, nil
	}

	return EnrichProfile200JSONResponse{
		Name:  pii.Name,
		Phone: pii.Phone,
		Address: Address{
			StreetAddress: pii.Address.StreetAddress,
			Locality:      pii.Address.Locality,
			Region:        pii.Address.Region,
			PostalCode:    pii.Address.PostalCode,
			Country:       pii.Address.Country,
		},
	}, nil
}

// NewRouter wires OpenAPI handlers and JWT bearer authentication middleware.
func NewRouter(s store.DAO, a *auth.Service, i idp.Connector) http.Handler {
	srv := NewServer(s, a, i)
	strictHandler := NewStrictHandler(srv, nil)
	apiHandler := Handler(strictHandler)

	// Auth gate middleware: enforces JWT Bearer token on protected routes (/profiles)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, "/profiles") {
			a.Middleware(apiHandler).ServeHTTP(w, r)
			return
		}
		// Public routes (/auth/login, /auth/register, health, etc.)
		apiHandler.ServeHTTP(w, r)
	})
}
