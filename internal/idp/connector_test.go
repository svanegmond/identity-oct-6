package idp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/svanegmond/agentic-eng-oct-6/internal/idp"
)

func TestIdPConnector_WireMappingAndBothConfigs(t *testing.T) {
	var authCalls int32
	var identityCalls int32

	// Mock vendor server implementing /auth and /identity
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth":
			atomic.AddInt32(&authCalls, 1)
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			if body["username"] != "vendor_user" || body["password"] != "vendor_secret" {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid_credentials"})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "vendor-jwt-token-xyz",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})

		case "/identity":
			atomic.AddInt32(&identityCalls, 1)
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer vendor-jwt-token-xyz" {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
				return
			}
			var req map[string]string
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"name":  req["name"],
				"phone": req["phone"],
				"address": map[string]string{
					"street_address": "500 Howard St",
					"locality":       "San Francisco",
					"region":         "CA",
					"postal_code":    "94105",
					"country":        "USA",
				},
			})

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	ctx := context.Background()

	// Test 1: Config ABC
	cfgABC := idp.ProviderConfig{
		Name:     "ABC",
		BaseURL:  server.URL,
		Username: "vendor_user",
		Password: "vendor_secret",
		Timeout:  5 * time.Second,
	}
	clientABC := idp.NewClient(cfgABC)

	// Direct Authenticate test
	token, err := clientABC.Authenticate(ctx, "vendor_user", "vendor_secret")
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}
	if token != "vendor-jwt-token-xyz" {
		t.Fatalf("unexpected token: %s", token)
	}

	// Direct GetIdentity test
	pii, err := clientABC.GetIdentity(ctx, token, "John Doe", "+15551234567")
	if err != nil {
		t.Fatalf("GetIdentity failed: %v", err)
	}
	if pii.Name != "John Doe" || pii.Phone != "+15551234567" {
		t.Errorf("PII mismatch: got %+v", pii)
	}
	if pii.Address.StreetAddress != "500 Howard St" ||
		pii.Address.Locality != "San Francisco" ||
		pii.Address.Region != "CA" ||
		pii.Address.PostalCode != "94105" ||
		pii.Address.Country != "USA" {
		t.Errorf("Address mapping mismatch: got %+v", pii.Address)
	}

	// Test 2: FetchIdentity with automatic auth & token caching
	piiCached, err := clientABC.FetchIdentity(ctx, "Jane Roe", "+15559876543")
	if err != nil {
		t.Fatalf("FetchIdentity failed: %v", err)
	}
	if piiCached.Name != "Jane Roe" {
		t.Errorf("FetchIdentity name mismatch: got %s", piiCached.Name)
	}

	// Second FetchIdentity should use cached token, so authCalls should not increment
	authCountBefore := atomic.LoadInt32(&authCalls)
	_, err = clientABC.FetchIdentity(ctx, "Jane Roe 2", "+15559876543")
	if err != nil {
		t.Fatalf("Second FetchIdentity failed: %v", err)
	}
	authCountAfter := atomic.LoadInt32(&authCalls)
	if authCountAfter != authCountBefore {
		t.Errorf("expected cached token to be used without calling /auth again; before=%d, after=%d", authCountBefore, authCountAfter)
	}

	// Test 3: Config XYC targeting same interface with separate config
	cfgXYC := idp.ProviderConfig{
		Name:     "XYC",
		BaseURL:  server.URL,
		Username: "vendor_user",
		Password: "vendor_secret",
		Timeout:  5 * time.Second,
	}
	clientXYC := idp.NewClient(cfgXYC)
	piiXYC, err := clientXYC.FetchIdentity(ctx, "Bob Smith", "+15550001111")
	if err != nil {
		t.Fatalf("FetchIdentity on XYC client failed: %v", err)
	}
	if piiXYC.Name != "Bob Smith" {
		t.Errorf("XYC PII mismatch: got %s", piiXYC.Name)
	}
}
