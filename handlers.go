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

	return router
}

// registerUser handle the http request to register a
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
