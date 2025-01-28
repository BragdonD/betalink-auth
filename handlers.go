package betalinkauth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/BragdonD/betalink-auth/middleware"
	betalinklogger "github.com/BragdonD/betalink-logger"
	"github.com/gin-gonic/gin"
)

// registerUserDto is the data transfer object for registering a new user
type registerUserDto struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// loginUserDto is the data transfer object for logging in a user
type loginUserDto struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Router is the http router for the auth service
type Router struct {
	logger   *betalinklogger.Logger
	router   *gin.Engine
	usecases *Usecases
}

// NewRouter creates a new Router instance
func NewRouter(logger *betalinklogger.Logger, ginRouter *gin.Engine, usecases *Usecases) *Router {
	router := &Router{
		logger:   logger,
		router:   ginRouter,
		usecases: usecases,
	}

	ginRouter.POST("/register", router.registerUser)
	ginRouter.POST("/login", router.loginUser)
	ginRouter.GET("/token/validate", router.validateAccessToken)
	ginRouter.GET("/token/refresh", router.refreshToken)

	return router
}

// registerUser handles the http request to register a
// new user in the database
func (r *Router) registerUser(ctx *gin.Context) {
	r.logger.Info("Registering user")
	var dto registerUserDto
	if err := ctx.BindJSON(&dto); err != nil {
		writeResponse(
			ctx,
			http.StatusBadRequest,
			false,
			nil,
			fmt.Errorf("could not bind json: %w", err),
		)
		return
	}
	if err := r.usecases.RegisterUser(
		ctx, dto.FirstName, dto.LastName, dto.Email, dto.Password); err != nil {
		writeResponse(
			ctx,
			http.StatusBadRequest,
			false,
			nil,
			fmt.Errorf("could not register the user: %w", err),
		)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "user registered"})
}

// loginUser handles the http request to login a user
func (r *Router) loginUser(ctx *gin.Context) {
	r.logger.Info("Logging in user")
	var dto loginUserDto
	if err := ctx.BindJSON(&dto); err != nil {
		writeResponse(
			ctx,
			http.StatusBadRequest,
			false,
			nil,
			fmt.Errorf("could not bind json: %w", err),
		)
		return
	}
	tokens, err := r.usecases.LoginUser(ctx, dto.Email, dto.Password)
	if err != nil {
		statusCode := getErrorStatusCode(err)
		writeResponse(
			ctx,
			statusCode,
			false,
			nil,
			fmt.Errorf("could not login the user: %w", err),
		)
		return
	}

	if ctx.Writer.Header().Get("Authorization") == "" {
		ctx.Writer.Header().Add("Authorization", "Bearer "+tokens.AccessToken)
	}
	ctx.SetCookie("refresh_token", tokens.RefreshToken, 3600, "/", "localhost", false, true)
	writeResponse(ctx, http.StatusOK, true, nil, nil)
}

// validateAccessToken handles the http request to validate an access token
func (r *Router) validateAccessToken(ctx *gin.Context) {
	r.logger.Info("Validating access token")
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		writeResponse(
			ctx,
			http.StatusUnauthorized,
			false,
			nil,
			fmt.Errorf("authorization header is required"),
		)
		return
	}
	// Validate the format of the Authorization header
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		writeResponse(
			ctx,
			http.StatusUnauthorized,
			false,
			nil,
			fmt.Errorf("invalid Authorization header format"),
		)
		return
	}
	accessToken := parts[1]

	user, err := r.usecases.ValidateAccessToken(ctx, accessToken)
	if err != nil {
		if err == ExpiredTokenError {
			// redirect to refresh token endpoint
			ctx.Redirect(http.StatusUnauthorized, "/token/refresh")
			return
		}

		statusCode := getErrorStatusCode(err)
		writeResponse(
			ctx,
			statusCode,
			false,
			nil,
			fmt.Errorf("could not validate access token: %w", err),
		)
		return
	}

	userdata := middleware.UserData{
		UserID:    user.UserID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	writeResponse(ctx, http.StatusOK, true, userdata, nil)
}

// refreshToken handles the http request to refresh an access token
func (r *Router) refreshToken(ctx *gin.Context) {
	r.logger.Info("Refreshing access token")
	// Get the refresh token from the cookies
	cookies := ctx.Request.Cookies()
	var refreshToken string
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			refreshToken = cookie.Value
			break
		}
	}
	if refreshToken == "" {
		writeResponse(
			ctx,
			http.StatusUnauthorized,
			false,
			nil,
			fmt.Errorf("refresh token is required"),
		)
		return
	}

	tokens, err := r.usecases.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		statusCode := getErrorStatusCode(err)
		writeResponse(
			ctx,
			statusCode,
			false,
			nil,
			fmt.Errorf("could not refresh access token: %w", err),
		)
		return
	}

	ctx.Writer.Header().Add("Authorization", "Bearer "+tokens.AccessToken)
	writeResponse(ctx, http.StatusOK, true, nil, nil)
}

// getErrorStatusCode returns the status code for an error
func getErrorStatusCode(err error) int {
	switch err.(type) {
	case *ValidationError:
		return http.StatusBadRequest
	case *ServerError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// writeResponse writes a response to the client
func writeResponse(ctx *gin.Context, status int, success bool, data interface{}, err error) {
	var errStr string
	if err != nil {
		errStr = err.Error()
	}

	ctx.JSON(status, gin.H{
		"success": success,
		"data":    data,
		"error":   errStr,
	})
}
