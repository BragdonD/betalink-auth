package betalinkauth

import (
	"context"
	"fmt"

	betalinklogger "github.com/BragdonD/betalink-logger"
)

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

	// check email uniqueness

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

	return &IDTokens{
		AccessToken:  accessToken,
		RefreshToken: "",
	}, nil
}
