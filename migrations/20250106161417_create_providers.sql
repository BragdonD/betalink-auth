-- +goose Up

INSERT INTO ExternalLoginProviders (provider_name, provider_endpoint) VALUES ('Google', 'https://accounts.google.com/o/oauth2/v2/auth');
