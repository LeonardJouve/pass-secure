CREATE TABLE IF NOT EXISTS users (
    id   BIGSERIAL PRIMARY KEY,
    email VARCHAR(128) UNIQUE NOT NULL,
    username VARCHAR(128) UNIQUE NOT NULL,
	password VARCHAR(512) NOT NULL
);

CREATE TABLE IF NOT EXISTS folders (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    owner_id BIGINT NOT NULL,
    parent_id BIGINT NULL,
    CONSTRAINT folders_owner_fk FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT folders_parent_fk FOREIGN KEY (parent_id) REFERENCES folders(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS entries (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    username VARCHAR(128) NOT NULL,
    password VARCHAR(512) NOT NULL,
    url VARCHAR(512) NULL,
    folder_id BIGINT NOT NULL,
    CONSTRAINT entries_folder_fk FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_folders (
    user_id BIGINT,
    folder_id BIGINT,
    CONSTRAINT user_folders_pk PRIMARY KEY (user_id, folder_id),
    CONSTRAINT user_folders_user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT user_folders_folder_fk FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE
);

CREATE OR REPLACE FUNCTION create_root_folder()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO folders(owner_id, name, parent_id)
    VALUES(NEW.id, '', NULL);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER create_root_folder
AFTER INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION create_root_folder();

CREATE OR REPLACE FUNCTION create_user_folder()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO user_folders(user_id, folder_id)
    VALUES(NEW.owner_id, NEW.id);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER create_user_folder
AFTER INSERT ON folders
FOR EACH ROW
EXECUTE FUNCTION create_user_folder();

CREATE OR REPLACE FUNCTION create_owner_user_folder()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO user_folders (user_id, folder_id)
    VALUES (NEW.owner_id, NEW.id)
    ON CONFLICT (user_id, folder_id) DO NOTHING;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER create_owner_user_folder
AFTER UPDATE ON folders
FOR EACH ROW
EXECUTE FUNCTION create_owner_user_folder();

CREATE OR REPLACE FUNCTION test()
RETURNS trigger AS $$
BEGIN
    PERFORM pg_notify(
        'websocket_events',
        json_build_object(
            'table', TG_TABLE_NAME,
            'schema', TG_TABLE_SCHEMA,
            'operation', TG_OP,
            'id', NEW.id
        )::text
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER test
AFTER UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION test();
