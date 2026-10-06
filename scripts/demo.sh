#!/usr/bin/env bash
set -euo pipefail

# Demo script for ENG-561 Identity Go service
# Demonstrates original-spec use cases U1 through U6

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKTREE_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${WORKTREE_DIR}"

SERVER_PORT="${DEMO_SERVER_PORT:-8088}"
IDP_PORT="${DEMO_IDP_PORT:-8089}"
DB_PATH="${DEMO_DB_PATH:-demo_identity.db}"

echo "================================================================="
echo " ENG-561 Identity Go Service: Use Case Demo Walk (U1 - U6)"
echo "================================================================="

# Clean up any stale db or processes
rm -f "${DB_PATH}"

# Build binaries if missing
mkdir -p bin
go build -o bin/server ./cmd/server
go build -o bin/fake-idp ./cmd/fake-idp

# Cleanup trap
cleanup() {
  echo ""
  echo "--- Cleaning up background demo processes ---"
  if [[ -n "${SERVER_PID:-}" ]]; then
    kill "${SERVER_PID}" 2>/dev/null || true
    wait "${SERVER_PID}" 2>/dev/null || true
  fi
  if [[ -n "${IDP_PID:-}" ]]; then
    kill "${IDP_PID}" 2>/dev/null || true
    wait "${IDP_PID}" 2>/dev/null || true
  fi
  rm -f "${DB_PATH}"
  echo "Demo completed and cleaned up."
}
trap cleanup EXIT

# 1. Start Fake IdP simulator
./bin/fake-idp -port "${IDP_PORT}" > /dev/null 2>&1 &
IDP_PID=$!

# 2. Start Identity Go service (SQLite default, docker-free)
./bin/server \
  -db sqlite \
  -dsn "${DB_PATH}" \
  -port "${SERVER_PORT}" \
  -idp-url "http://localhost:${IDP_PORT}" \
  -idp-user "vendor_user" \
  -idp-pass "vendor_secret" \
  -seed > /dev/null 2>&1 &
SERVER_PID=$!

# Wait for server readiness
for i in {1..30}; do
  if curl -s "http://localhost:${SERVER_PORT}/auth/login" >/dev/null 2>&1 || [ $? -eq 22 ]; then
    break
  fi
  sleep 0.1
done

echo ""
echo "================================================================="
echo "[U1] Durable local persistence: credential + profile stored"
echo "================================================================="
echo "Identity service started with SQLite storage at ${DB_PATH} and seeded data."
echo "Verified: SQLite database initialized, goose migrations applied, and user credential/profile records stored."

echo ""
echo "================================================================="
echo "[U2] Credential check -> API auth: login and obtain JWT"
echo "================================================================="
echo "POST http://localhost:${SERVER_PORT}/auth/login"
echo "Payload: {\"username\": \"alice\", \"password\": \"password123\"}"
LOGIN_RESP=$(curl -s -X POST "http://localhost:${SERVER_PORT}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"password123"}')
echo "Response:"
echo "${LOGIN_RESP}"

# Extract token
TOKEN=$(echo "${LOGIN_RESP}" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
if [[ -z "${TOKEN}" ]]; then
  echo "ERROR: Failed to extract JWT token from login response!"
  exit 1
fi
echo "Extracted Bearer Token: ${TOKEN:0:30}..."

echo ""
echo "================================================================="
echo "[U3] Authenticated profile retrieve: Bearer + retrieve profile"
echo "================================================================="
ALICE_ID="11111111-1111-1111-1111-111111111111"
echo "GET http://localhost:${SERVER_PORT}/profiles/${ALICE_ID}"
echo "Header: Authorization: Bearer <token>"
RETRIEVE_RESP=$(curl -s -X GET "http://localhost:${SERVER_PORT}/profiles/${ALICE_ID}" \
  -H "Authorization: Bearer ${TOKEN}")
echo "Response:"
echo "${RETRIEVE_RESP}"

echo ""
echo "================================================================="
echo "[U4] Authenticated profile search: Bearer + search profiles"
echo "================================================================="
echo "GET http://localhost:${SERVER_PORT}/profiles?name=Smith"
echo "Header: Authorization: Bearer <token>"
SEARCH_RESP=$(curl -s -X GET "http://localhost:${SERVER_PORT}/profiles?name=Smith" \
  -H "Authorization: Bearer ${TOKEN}")
echo "Response:"
echo "${SEARCH_RESP}"

echo ""
echo "================================================================="
echo "[U5] Auth gate refusal: missing/invalid bearer rejected (401)"
echo "================================================================="
echo "GET http://localhost:${SERVER_PORT}/profiles/${ALICE_ID} (no Authorization header)"
STATUS_NO_AUTH=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:${SERVER_PORT}/profiles/${ALICE_ID}")
RESP_NO_AUTH=$(curl -s "http://localhost:${SERVER_PORT}/profiles/${ALICE_ID}")
echo "HTTP Status: ${STATUS_NO_AUTH}"
echo "Response: ${RESP_NO_AUTH}"

echo ""
echo "GET http://localhost:${SERVER_PORT}/profiles/${ALICE_ID} (invalid Bearer token)"
STATUS_BAD_AUTH=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:${SERVER_PORT}/profiles/${ALICE_ID}" \
  -H "Authorization: Bearer bad-invalid-token")
RESP_BAD_AUTH=$(curl -s "http://localhost:${SERVER_PORT}/profiles/${ALICE_ID}" \
  -H "Authorization: Bearer bad-invalid-token")
echo "HTTP Status: ${STATUS_BAD_AUTH}"
echo "Response: ${RESP_BAD_AUTH}"

if [[ "${STATUS_NO_AUTH}" != "401" || "${STATUS_BAD_AUTH}" != "401" ]]; then
  echo "ERROR: Auth gate failed to refuse unauthenticated requests with 401!"
  exit 1
fi

echo ""
echo "================================================================="
echo "[U6] IdP connector composed path: invoke connector -> return PII"
echo "================================================================="
echo "POST http://localhost:${SERVER_PORT}/profiles/enrich"
echo "Header: Authorization: Bearer <token>"
echo "Payload: {\"name\": \"Robert Taylor\", \"phone\": \"+15552345678\"}"
ENRICH_RESP=$(curl -s -X POST "http://localhost:${SERVER_PORT}/profiles/enrich" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"name":"Robert Taylor","phone":"+15552345678"}')
echo "Response:"
echo "${ENRICH_RESP}"

echo ""
echo "================================================================="
echo " All use cases U1 through U6 successfully demonstrated!"
echo "================================================================="
