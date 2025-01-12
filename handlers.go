package betalinkauth

import (
	"fmt"
	"net/http"

	betalinklogger "github.com/BragdonD/betalink-logger"
	"github.com/gin-gonic/gin"
)

// registerUserDto is the data transfer object for registering a new user
type registerUserDto struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
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

	return router
}

// registerUser handles the http request to register a
// new user in the database
func (r *Router) registerUser(ctx *gin.Context) {
	r.logger.Info("Registering user")
	var dto registerUserDto
	if err := ctx.BindJSON(&dto); err != nil {
		writeError(ctx, fmt.Errorf("could not bind json: %w", err))
		return
	}
	if err := r.usecases.RegisterUser(
		ctx, dto.FirstName, dto.LastName, dto.Email, dto.Password); err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "user registered"})
}

// loginUser handles the http request to login a user
func (r *Router) loginUser(ctx *gin.Context) {
	r.logger.Info("Logging in user")
	var dto loginUserDto
	if err := ctx.BindJSON(&dto); err != nil {
		writeError(ctx, fmt.Errorf("could not bind json: %w", err))
		return
	}
	tokens, err := r.usecases.LoginUser(ctx, dto.Email, dto.Password)
	if err != nil {
		writeError(ctx, err)
		return
	}

	// Add tokens to the header
	addToHeader(ctx.Writer.Header(), "Authorization", "Bearer "+tokens.AccessToken)
	addToHeader(ctx.Writer.Header(), "Refresh-Token", tokens.RefreshToken)

	writeResponse(ctx, http.StatusOK, gin.H{"message": "user logged in"})
}

// addToHeader adds a key-value pair to the header
func addToHeader(header http.Header, key, value string) {
	if header.Get(key) == "" {
		header.Add(key, value)
	}
}

// writeError writes an error response to the client
func writeError(ctx *gin.Context, err error) {
	switch e := err.(type) {
	case *ValidationError:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": e.Error(),
		})
	case *ServerError:
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": e.Error(),
		})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
}

// writeResponse writes a response to the client
func writeResponse(ctx *gin.Context, status int, data interface{}) {
	ctx.JSON(status, data)
}
