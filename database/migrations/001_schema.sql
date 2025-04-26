CREATE TABLE IF NOT EXISTS users (
    id   BIGSERIAL PRIMARY KEY,
    email VARCHAR(100) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
	password VARCHAR(100) NOT NULL
);

CREATE TABLE IF NOT EXISTS folders (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50),
    owner_id BIGINT NOT NULL,
    parent_id BIGINT NULL,
    CONSTRAINT folders_owner_fk FOREIGN KEY (owner_id) REFERENCES users(id),
    CONSTRAINT folders_parent_fk FOREIGN KEY (parent_id) REFERENCES folders(id)
);

CREATE TABLE IF NOT EXISTS entries (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    password VARCHAR(100) NOT NULL,
    folder_id BIGINT NOT NULL,
    CONSTRAINT entries_folder_fk FOREIGN KEY (folder_id) REFERENCES folders(id)
);

CREATE TABLE IF NOT EXISTS user_folders (
    user_id BIGINT,
    folder_id BIGINT,
    CONSTRAINT user_folders_pk PRIMARY KEY (user_id, folder_id),
    CONSTRAINT user_folders_user_fk FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT user_folders_folder_fk FOREIGN KEY (folder_id) REFERENCES folders(id)
);

CREATE TABLE IF NOT EXISTS folder_entries (
    folder_id BIGINT,
    entry_id BIGINT,
    CONSTRAINT folder_entries_pk PRIMARY KEY (folder_id, entry_id),
    CONSTRAINT folder_entries_folder_fk FOREIGN KEY (folder_id) REFERENCES folders(id),
    CONSTRAINT folder_entries_entry_fk FOREIGN KEY (entry_id) REFERENCES entries(id)
);

-- TODO: delete cascading
