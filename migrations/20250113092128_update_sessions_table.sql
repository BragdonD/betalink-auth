-- +goose Up

ALTER TABLE Sessions
RENAME COLUMN session_token TO session_id;

ALTER TABLE Sessions
ALTER COLUMN session_id SET DATA TYPE UUID USING session_id::UUID;

ALTER TABLE Sessions
ALTER COLUMN session_id SET DEFAULT uuid_generate_v4();