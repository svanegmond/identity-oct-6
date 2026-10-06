-- name: GetProfileByID :one
SELECT id, name, address, phone, created_at, updated_at
FROM user_profiles
WHERE id = ?;

-- name: CreateProfile :exec
INSERT INTO user_profiles (id, name, address, phone, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: SearchProfiles :many
SELECT id, name, address, phone, created_at, updated_at
FROM user_profiles
WHERE (cast(@name as text) = '' OR name LIKE '%' || cast(@name as text) || '%')
  AND (cast(@phone as text) = '' OR phone LIKE '%' || cast(@phone as text) || '%')
ORDER BY created_at ASC;

-- name: GetCredentialByUsername :one
SELECT id, user_id, username, method, password, created_at, updated_at
FROM user_credentials
WHERE username = ?;

-- name: CreateCredential :exec
INSERT INTO user_credentials (id, user_id, username, method, password, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?);
