-- name: CreateVerifyEmail :one
INSERT INTO
    verify_emails (username, email, secret_code)
VALUES ($1, $2, $3)
RETURNING
    *;

-- name: UpdateVerifyEmail :one
UPDATE verify_emails
SET
    is_used = TRUE
WHERE
    id = $1
    AND secret_code = $2
    AND is_used = FALSE
    AND expired_at > now()
RETURNING
    *;

-- name: DeleteVerifyEmail :exec
DELETE FROM verify_emails WHERE id = $1;

-- name: GetVerifyEmail :one
SELECT * FROM verify_emails WHERE id = $1 LIMIT 1;

-- name: ListAllVerifyEmails :many
SELECT * FROM verify_emails ORDER BY id LIMIT $1 OFFSET $2;

-- name: DeleteAllVerifyEmails :exec
DELETE FROM verify_emails;