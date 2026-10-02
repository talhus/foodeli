package auth

import (
	"errors"
	"net/http"

	"foodeli/platform/middleware"
	"foodeli/platform/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc, rateLimiters ...gin.HandlerFunc) {
	var loginLimiter, registerLimiter gin.HandlerFunc
	if len(rateLimiters) > 0 {
		loginLimiter = rateLimiters[0]
	}
	if len(rateLimiters) > 1 {
		registerLimiter = rateLimiters[1]
	}

	authGroup := r.Group("/auth")
	{
		registerHandlers := []gin.HandlerFunc{}
		if registerLimiter != nil {
			registerHandlers = append(registerHandlers, registerLimiter)
		}
		registerHandlers = append(registerHandlers, h.Register)
		authGroup.POST("/register", registerHandlers...)

		loginHandlers := []gin.HandlerFunc{}
		if loginLimiter != nil {
			loginHandlers = append(loginHandlers, loginLimiter)
		}
		loginHandlers = append(loginHandlers, h.Login)
		authGroup.POST("/login", loginHandlers...)

		authGroup.GET("/me", authMiddleware, h.Me)
		authGroup.POST("/logout", h.Logout)
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new customer or merchant account
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration credentials"
// @Success 201 {object} response.Response{data=AuthResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	res, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			response.Conflict(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, res)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and issue JWT accessToken cookie and bearer token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} response.Response{data=AuthResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid credentials payload: "+err.Error())
		return
	}

	res, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Unauthorized(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to login: "+err.Error())
		return
	}

	c.SetCookie("accessToken", res.Token, 86400, "/", "", false, true)

	response.Success(c, http.StatusOK, res)
}

// Me godoc
// @Summary Get authenticated user profile
// @Description Retrieve profile of currently authenticated user
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=User}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	profile, err := h.service.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.NotFound(c, "User profile not found")
			return
		}
		response.InternalServerError(c, "Failed to fetch profile: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, profile)
}

// Logout godoc
// @Summary Logout user
// @Description Clear authentication session cookie
// @Tags Auth
// @Produce json
// @Success 200 {object} response.Response
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	c.SetCookie("accessToken", "", -1, "/", "", false, true)
	response.Success(c, http.StatusOK, "Logout successful")
}
