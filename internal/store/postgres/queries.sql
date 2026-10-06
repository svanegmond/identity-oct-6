-- name: GetProfileByID :one
SELECT id, name, address, phone, created_at, updated_at
FROM user_profiles
WHERE id = $1;

-- name: CreateProfile :exec
INSERT INTO user_profiles (id, name, address, phone, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: SearchProfiles :many
SELECT id, name, address, phone, created_at, updated_at
FROM user_profiles
WHERE (cast(@name as text) = '' OR name ILIKE '%' || cast(@name as text) || '%')
  AND (cast(@phone as text) = '' OR phone ILIKE '%' || cast(@phone as text) || '%')
ORDER BY created_at ASC;

-- name: GetCredentialByUsername :one
SELECT id, user_id, username, method, password, created_at, updated_at
FROM user_credentials
WHERE username = $1;

-- name: CreateCredential :exec
INSERT INTO user_credentials (id, user_id, username, method, password, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);
