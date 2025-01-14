package betalinkauth

import (
	"context"
	"fmt"
	"reflect"
	"time"

	betalinklogger "github.com/BragdonD/betalink-logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// UserData represents the user information retrieved from the auth server
type UserData struct {
	UserID    pgtype.UUID
	FirstName string
	LastName  string
}

// IDTokens is a struct containing the access and refresh tokens
type IDTokens struct {
	AccessToken  string
	RefreshToken string
}

// Usecases is the usecases for the auth service
type Usecases struct {
	logger  *betalinklogger.Logger
	queries *Queries
}

// NewUsecase creates a new Usecases instance
func NewUsecase(logger *betalinklogger.Logger, queries *Queries) *Usecases {
	return &Usecases{
		logger:  logger,
		queries: queries,
	}
}

// RegisterUser registers a new user in the database
func (u *Usecases) RegisterUser(ctx context.Context, firstname, lastname, email, password string) error {
	u.logger.Info("Registering user")
	// validate user data
	if ok, err := ValidateEmail(email); !ok {
		return &ValidationError{
			Message: fmt.Errorf("could not validate email: %w", err).Error(),
		}
	}
	if ok, err := ValidatePassword(password); !ok {
		return &ValidationError{
			Message: fmt.Errorf("could not validate password: %w", err).Error(),
		}
	}

	err := checkEmailUniqueness(ctx, u.queries, email)
	if err != nil {
		return err
	}

	// create user
	userParams := CreateUserParams{
		FirstName: firstname,
		LastName:  lastname,
	}
	userID, err := u.queries.CreateUser(ctx, userParams)
	if err != nil {
		return &ServerError{
			Message: fmt.Errorf("could not create user: %w", err).Error(),
		}
	}

	// create user login data
	passwordHash, err := HashPassword(password)
	if err != nil {
		return &ServerError{
			Message: fmt.Errorf("could not hash password: %w", err).Error(),
		}
	}
	userLoginDataParams := CreateUserLoginDataParams{
		UserID:        userID,
		Email:         email,
		Passwordhash:  passwordHash,
		Passwordsalt:  "",
		Hashalgorithm: "BCRYPT",
	}
	err = u.queries.CreateUserLoginData(ctx, userLoginDataParams)
	if err != nil {
		return &ServerError{
			Message: fmt.Errorf("could not create user login data: %w", err).Error(),
		}
	}

	// create email verification
	// TODO: implement email verification

	return nil
}

// checkEmailUniqueness checks if an email is unique in the database
func checkEmailUniqueness(ctx context.Context, queries *Queries, email string) error {
	// Attempt to get login data by email
	_, err := queries.GetLoginDataByEmail(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Email does not exist; it's unique
			return nil
		}
		fmt.Println(reflect.TypeOf(err))
		// Handle other unexpected errors
		return &ServerError{
			Message: fmt.Errorf("could not get login data by email: %w", err).Error(),
		}
	}

	// Email already exists
	return &ValidationError{
		Message: fmt.Sprintf("email [%s] is not available", email),
	}
}

// LoginUser checks the user credentials
func (u *Usecases) LoginUser(ctx context.Context, email, password string) (*IDTokens, error) {
	u.logger.Info("Logging in user")
	// get login data
	loginData, err := u.queries.GetLoginDataByEmail(ctx, email)
	if err != nil {
		return nil, &ServerError{
			Message: fmt.Errorf("could not get login data: %w", err).Error(),
		}
	}

	// check password
	if err := ComparePassword(password, loginData.Passwordhash); err != nil {
		return nil, &ValidationError{
			Message: fmt.Errorf("could not compare password: %w", err).Error(),
		}
	}

	// create refresh and access tokens
	// TODO: implement roles
	// TODO: implement secret
	accessToken, err := GenerateAccessToken(loginData.UserID.String(), []string{"user"}, "mysecret")
	if err != nil {
		return nil, &ServerError{
			Message: fmt.Errorf("could not generate access token: %w", err).Error(),
		}
	}

	createSessionParams := CreateSessionParams{
		UserID: loginData.UserID,
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour * 24), // TODO: make this configurable
			Valid: true,
		},
	}
	sessionID, err := u.queries.CreateSession(ctx, createSessionParams)
	if err != nil {
		return nil, &ServerError{
			Message: fmt.Errorf("could not create session: %w", err).Error(),
		}
	}
	refreshToken, err := GenerateRefreshToken(
		sessionID.String(),
		createSessionParams.CreatedAt.Time,
		createSessionParams.ExpiresAt.Time,
		"mysecret",
	)
	if err != nil {
		return nil, &ServerError{
			Message: fmt.Errorf("could not generate refresh token: %w", err).Error(),
		}
	}

	return &IDTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateAccessToken validates an access token
func (u *Usecases) ValidateAccessToken(ctx context.Context, accessToken string) (*UserData, error) {
	// validate access token
	claims, err := ValidateAccessToken(accessToken, "mysecret")
	if err != nil {
		return nil, &ValidationError{
			Message: fmt.Errorf("could not validate access token: %w", err).Error(),
		}
	}

	expiresAt, err := claims.GetExpirationTime()
	if err != nil {
		return nil, &ValidationError{
			Message: fmt.Errorf("could not get expiration time: %w", err).Error(),
		}
	}

	// check if token is expired
	if expiresAt.Time.Before(time.Now()) {
		return nil, ExpiredTokenError
	}

	// get user ID
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, &ValidationError{
			Message: "could not get user ID from claims",
		}
	}
	parsedUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, &ValidationError{
			Message: "invalid UUID format",
		}
	}
	pgUUID := pgtype.UUID{
		Bytes: parsedUUID,
		Valid: true,
	}

	user, err := u.queries.GetUserById(ctx, pgUUID)
	if err != nil {
		return nil, fmt.Errorf("could not get user by ID: %w", err)
	}

	return &UserData{
		UserID:    user.UserID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, nil
}
