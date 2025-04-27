-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByEmailOrUsername :one
SELECT * FROM users
WHERE email = $1 OR username = $2 LIMIT 1;

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
    SELECT entry_id FROM folder_entries
    WHERE folder_id IN (
        SELECT folder_id FROM user_folders WHERE user_id = $1
    )
);

-- name: GetUserEntry :one
SELECT * FROM entries
WHERE id IN (
    SELECT entry_id FROM folder_entries
    WHERE folder_id IN (
        SELECT folder_id FROM user_folders WHERE user_id = $1
    )
) AND id = sqlc.arg(entry_id);

-- name: CreateEntry :one
WITH entry AS (
    INSERT INTO entries(name, username, password, url, folder_id)
    VALUES($1, $2, $3, $4, $5)
    RETURNING *
),
folder_entry AS (
    INSERT INTO folder_entries(folder_id, entry_id)
    SELECT folder_id, id FROM entry
    RETURNING *
)
SELECT * FROM entries
WHERE id = (
    SELECT id FROM entry
);

-- name: UpdateEntry :one
WITH old_entry AS (
    SELECT * FROM entries
    WHERE id = $1
),
new_entry AS (
    UPDATE entries
    SET name = $2, username = $3, password = $4, url = $5, folder_id = $6
    WHERE id = $1
    RETURNING *
),
folder_entry AS (
    UPDATE folder_entries
    SET folder_id = $6
    WHERE folder_id = (
        SELECT folder_id FROM old_entry
    ) AND entry_id = $1
)
SELECT * FROM entries
WHERE entries.id = $1;

-- name: DeleteEntry :exec
DELETE FROM entries
WHERE id = $1;

-- name: GetUserRootFolder :one
SELECT * FROM folders
WHERE parent_id = NULL AND id IN (
    SELECT folder_id FROM user_folders WHERE user_id = $1
);

-- name: CreateFolder :one
WITH folder AS (
    INSERT INTO folders(owner_id, name, parent_id)
    VALUES($1, $2, $3)
    RETURNING *
),
user_folder AS (
    INSERT INTO user_folders(user_id, folder_id)
    SELECT owner_id, id FROM folder
)
SELECT * FROM folders
WHERE id = (
    SELECT id FROM folder
);

-- name: UpdateFolder :one
WITH folder AS (
    UPDATE folders
    SET name = $2, owner_id = $3, parent_id = $4
    WHERE folders.id = $1
    RETURNING *
),
user_folder AS (
    INSERT INTO user_folders (user_id, folder_id)
    SELECT owner_id, id FROM folder
    ON CONFLICT (user_id, folder_id) DO NOTHING
    RETURNING *
)
SELECT * FROM folders
WHERE folders.id = (
    SELECT id FROM folder
);

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

