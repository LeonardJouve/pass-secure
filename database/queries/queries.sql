-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: HasUser :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE id = $1
) AS exists;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: HasUserWithEmailOrUsername :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE email = $1 OR username = $2
) AS exists;

-- name: CreateUser :one
INSERT INTO users(email, username, password)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET email = $2, username = $3, password = $4
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: GetUserEntries :many
SELECT * FROM entries
WHERE id IN (
    SELECT id FROM entries
    WHERE folder_id IN (
        SELECT folder_id FROM user_folders WHERE user_id = $1
    )
);

-- name: GetUserEntry :one
SELECT * FROM entries
WHERE entries.id = sqlc.arg(entry_id) AND entries.id IN (
    SELECT id FROM entries
    WHERE folder_id IN (
        SELECT folder_id FROM user_folders WHERE user_id = $1
    )
);

-- name: CreateEntry :one
INSERT INTO entries(name, username, password, url, folder_id)
VALUES($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateEntry :one
UPDATE entries
SET name = $2, username = $3, password = $4, url = $5, folder_id = $6
WHERE id = $1
RETURNING *;

-- name: DeleteEntry :exec
DELETE FROM entries
WHERE id = $1;

-- name: GetUserRootFolder :one
SELECT * FROM folders
WHERE parent_id = NULL AND id IN (
    SELECT folder_id FROM user_folders WHERE user_id = $1
);

-- name: CreateFolder :one
INSERT INTO folders(owner_id, name, parent_id)
VALUES($1, $2, $3)
RETURNING *;

-- name: UpdateFolder :one
UPDATE folders
SET name = $2, owner_id = $3, parent_id = $4
WHERE folders.id = $1
RETURNING *;

-- name: DeleteFolder :exec
DELETE FROM folders
WHERE id = $1;

-- name: GetUserFolders :many
SELECT * FROM folders
WHERE id IN (
    SELECT folder_id FROM user_folders WHERE user_id = $1
);

-- name: GetUserFolder :one
SELECT * FROM folders
WHERE id IN (
    SELECT folder_id FROM user_folders WHERE user_id = $1
) AND id = sqlc.arg(folder_id);

-- name: GetFolder :one
SELECT * FROM folders
WHERE id = $1;

-- name: GetFolderUsers :many
SELECT user_id FROM user_folders
WHERE folder_id = $1;

-- name: GetFoldersUsers :many
SELECT * FROM user_folders
WHERE folder_id = ANY($1::bigint[]);

-- name: AddFolderUser :exec
INSERT INTO user_folders(user_id, folder_id)
VALUES($1, $2)
ON CONFLICT (user_id, folder_id) DO NOTHING;

-- name: DeleteFolderUser :exec
DELETE FROM user_folders
WHERE user_id = $1 AND folder_id = $2;
