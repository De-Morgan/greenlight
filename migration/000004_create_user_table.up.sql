CREATE TABLE IF NOT EXISTS users(
id bigserial PRIMARY KEY,
created_at timestamptz(0) Not NULL DEFAULT NOW(),
name text NOT NULL,
email citext UNIQUE NOT NULL,
password_hash bytea NOT NULL,
activated bool NOT NULL DEFAULT 'f',
version integer NOT NULL DEFAULT 1
);