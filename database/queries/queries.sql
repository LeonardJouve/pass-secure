-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users(email, username, password)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserFolders :many
SELECT * FROM folders
WHERE id IN (
    SELECT folder_id FROM user_folders WHERE user_id = $1
);

-- name: GetFolderEntries :many
SELECT * FROM entries
WHERE id IN (
    SELECT entry_id FROM folder_entries WHERE folder_entries.folder_id = $1
);

-- name: GetFolderUserIds :many
SELECT user_id AS id FROM user_folders
WHERE folder_id = $1;
