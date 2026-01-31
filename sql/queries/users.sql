-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING id, created_at, updated_at, email, is_chirpy_red;

-- name: GetUserPWHashByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: RefreshTokenNew :one
INSERT INTO refresh_tokens (token,created_at,updated_at,user_id,expires_at)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3
)
RETURNING *;

-- name: RefreshTokenGet :one
SELECT * FROM refresh_tokens
WHERE token = $1;

-- name: RefreshTokenRevoke :one
UPDATE refresh_tokens 
SET updated_at = NOW(), revoked_at = NOW()
WHERE token = $1
RETURNING *;

-- name: UpdateUserLogin :one
UPDATE users 
SET email = $2, hashed_password = $3, updated_at = NOW()
WHERE id = $1
RETURNING id, created_at, updated_at, email, is_chirpy_red;

-- name: AddRedtoUser :one
UPDATE users
SET updated_at = NOW(), is_chirpy_red = $2
WHERE id = $1
RETURNING id, created_at, updated_at, email, is_chirpy_red;