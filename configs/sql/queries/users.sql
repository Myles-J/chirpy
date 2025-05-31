-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;


-- name: UpdateUser :one
UPDATE users
SET
    updated_at = NOW(),
    email = $1,
    hashed_password = $2
WHERE id = $3
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserFromRefreshToken :one
SELECT rt.token, u.id, u.email FROM users u
JOIN refresh_tokens rt ON u.id = rt.user_id
WHERE rt.token = $1 AND rt.expires_at > NOW() AND rt.revoked_at IS NULL;


-- name: UpdateUserIsChirpyRed :one
UPDATE users
SET
    is_chirpy_red = $1
WHERE id = $2
RETURNING *;