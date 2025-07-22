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

-- name: FindUserByEmail :one
SELECT * FROM users
WHERE email = $1;


-- name: FindUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: UpdatedUser :one
UPDATE users
SET updated_at = NOW(),
    email = $1,
    hashed_password = $2
WHERE id = $3 RETURNING *;

-- name: UpdateUserChirpyRed :exec
UPDATE users
SET is_chirpy_red = $1
WHERE id = $2;
