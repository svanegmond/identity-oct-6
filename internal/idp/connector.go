package idp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var (
	ErrAuthenticationFailed = errors.New("idp authentication failed")
	ErrIdentityNotFound     = errors.New("identity not found in idp")
	ErrUpstreamFailure      = errors.New("idp upstream service error")
)

type Address struct {
	StreetAddress string `json:"street_address"`
	Locality      string `json:"locality"`
	Region        string `json:"region"`
	PostalCode    string `json:"postal_code"`
	Country       string `json:"country"`
}

type IdentityPII struct {
	Name    string  `json:"name"`
	Phone   string  `json:"phone"`
	Address Address `json:"address"`
}

type ProviderConfig struct {
	Name     string        // "ABC" or "XYC"
	BaseURL  string        // Provider base URL
	Username string        // Vendor client credentials
	Password string
	Timeout  time.Duration
}

type Connector interface {
	Authenticate(ctx context.Context, username, password string) (string, error)
	GetIdentity(ctx context.Context, token, name, phone string) (*IdentityPII, error)
	FetchIdentity(ctx context.Context, name, phone string) (*IdentityPII, error)
}

type Client struct {
	config     ProviderConfig
	httpClient *http.Client

	mu          sync.RWMutex
	cachedToken string
	tokenExpiry time.Time
}

func NewClient(cfg ProviderConfig) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Authenticate(ctx context.Context, username, password string) (string, error) {
	reqBody, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal auth request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.BaseURL+"/auth", bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: status %d", ErrAuthenticationFailed, resp.StatusCode)
	}

	var authResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", fmt.Errorf("failed to decode auth response: %w", err)
	}

	c.mu.Lock()
	c.cachedToken = authResp.AccessToken
	expSec := authResp.ExpiresIn
	if expSec <= 0 {
		expSec = 3600
	}
	c.tokenExpiry = time.Now().Add(time.Duration(expSec)*time.Second - 30*time.Second)
	c.mu.Unlock()

	return authResp.AccessToken, nil
}

func (c *Client) GetIdentity(ctx context.Context, token, name, phone string) (*IdentityPII, error) {
	reqBody, err := json.Marshal(map[string]string{
		"name":  name,
		"phone": phone,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal identity request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.BaseURL+"/identity", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create identity request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("identity request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrIdentityNotFound
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrAuthenticationFailed
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrUpstreamFailure, resp.StatusCode)
	}

	var pii IdentityPII
	if err := json.NewDecoder(resp.Body).Decode(&pii); err != nil {
		return nil, fmt.Errorf("failed to decode identity response: %w", err)
	}

	return &pii, nil
}

func (c *Client) FetchIdentity(ctx context.Context, name, phone string) (*IdentityPII, error) {
	token, err := c.getOrRefreshToken(ctx)
	if err != nil {
		return nil, err
	}

	pii, err := c.GetIdentity(ctx, token, name, phone)
	if err != nil && errors.Is(err, ErrAuthenticationFailed) {
		// Retry once with a fresh token
		c.mu.Lock()
		c.cachedToken = ""
		c.mu.Unlock()

		freshToken, refreshErr := c.getOrRefreshToken(ctx)
		if refreshErr != nil {
			return nil, refreshErr
		}
		return c.GetIdentity(ctx, freshToken, name, phone)
	}

	return pii, err
}

func (c *Client) getOrRefreshToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	if c.cachedToken != "" && time.Now().Before(c.tokenExpiry) {
		token := c.cachedToken
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	return c.Authenticate(ctx, c.config.Username, c.config.Password)
}
