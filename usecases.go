package betalinkauth

import (
	"context"
	"database/sql"
	"fmt"

	betalinklogger "github.com/BragdonD/betalink-logger"
)

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
		if err == sql.ErrNoRows {
			// Email does not exist; it's unique
			return nil
		}
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
