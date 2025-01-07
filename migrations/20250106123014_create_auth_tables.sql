-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE Users (
    user_id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE HashAlgorithm (
    hashAlgorithm VARCHAR(255) NOT NULL UNIQUE,
    PRIMARY KEY (hashAlgorithm)
);

CREATE TABLE UsersLoginData (
    user_id uuid PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    passwordHash VARCHAR(255) NOT NULL,
    passwordSalt VARCHAR(255) NOT NULL,
    hashAlgorithm VARCHAR(255) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES Users(user_id),
    FOREIGN KEY (hashAlgorithm) REFERENCES HashAlgorithm(hashAlgorithm)
);

CREATE TABLE PasswordRecovery (
    user_id uuid PRIMARY KEY,
    recovery_token VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES Users(user_id)
);

CREATE TABLE EmailVerification (
    user_id uuid PRIMARY KEY,
    verification_token VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES Users(user_id)
);

CREATE TABLE ExternalLoginProviders (
    provider_id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_name VARCHAR(255) NOT NULL,
    provider_endpoint TEXT NOT NULL
);

CREATE TABLE UserLoginExternal (
    user_id uuid PRIMARY KEY,
    provider_id uuid NOT NULL,
    provider_access_token VARCHAR(255) NOT NULL,
    provider_refresh_token VARCHAR(255) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES Users(user_id),
    FOREIGN KEY (provider_id) REFERENCES ExternalLoginProviders(provider_id)
);
-- +goose StatementEnd

