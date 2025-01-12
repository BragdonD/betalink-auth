-- name: CreateUser :one
INSERT INTO Users (first_name, last_name) VALUES ($1, $2) RETURNING user_id;

-- name: CreateUserLoginData :exec
INSERT INTO UsersLoginData (user_id, email, passwordHash, passwordSalt, hashAlgorithm) VALUES ($1, $2, $3, $4, $5);

-- name: CreatePasswordRecovery :exec
INSERT INTO PasswordRecovery (user_id, recovery_token) VALUES ($1, $2);

-- name: CreateEmailVerification :exec
INSERT INTO EmailVerification (user_id, verification_token) VALUES ($1, $2);

-- name: GetLoginDataByEmail :one
SELECT user_id, email, passwordHash, passwordSalt, hashAlgorithm FROM UsersLoginData WHERE email = $1;

-- name: GetUserById :one
SELECT user_id, first_name, last_name FROM Users WHERE user_id = $1;