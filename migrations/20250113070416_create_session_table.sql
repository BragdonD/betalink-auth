-- +goose Up

CREATE EXTENSION "pg_cron";

CREATE TABLE Sessions (
    session_token TEXT NOT NULL PRIMARY KEY,
    user_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    FOREIGN KEY (user_id) REFERENCES Users(user_id)
);

-- Delete expired sessions every 20 minutes
SELECT cron.schedule('session-cleanup', '20 * * * *', 'DELETE FROM Sessions WHERE expires_at < NOW()');