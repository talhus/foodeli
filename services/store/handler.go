package store

import (
	"errors"
	"math"
	"net/http"
	"strconv"

	"foodeli/services/auth"
	"foodeli/platform/middleware"
	"foodeli/platform/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler manages HTTP endpoints for stores and menu items.
type Handler struct {
	service Service
}

// NewHandler creates a new Handler for store and menu management.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers store and menu endpoints.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	stores := r.Group("/stores")
	{
		// Public store browsing
		stores.GET("", h.ListStores)
		stores.GET("/:id", h.GetStore)
		stores.GET("/:id/menu", h.ListMenuItems)

		// Merchant protected routes
		stores.POST("", authMiddleware, middleware.RequireRole(auth.RoleMerchant, auth.RoleAdmin), h.CreateStore)
		stores.GET("/merchant/mine", authMiddleware, middleware.RequireRole(auth.RoleMerchant, auth.RoleAdmin), h.GetMyStore)
		stores.POST("/:id/menu", authMiddleware, middleware.RequireRole(auth.RoleMerchant, auth.RoleAdmin), h.CreateMenuItem)
	}
}

// CreateStore godoc
// @Summary Create a store
// @Description Register a new merchant storefront (enforces 1 store per merchant)
// @Tags Stores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateStoreRequest true "Store details"
// @Success 201 {object} response.Response{data=Store}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /stores [post]
func (h *Handler) CreateStore(c *gin.Context) {
	merchantID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req CreateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid store payload: "+err.Error())
		return
	}

	store, err := h.service.CreateStore(c.Request.Context(), merchantID, &req)
	if err != nil {
		if errors.Is(err, ErrMerchantStoreAlreadyExists) {
			response.Conflict(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, store)
}

// ListStores godoc
// @Summary List stores
// @Description Retrieve list of all registered stores
// @Tags Stores
// @Produce json
// @Success 200 {object} response.Response{data=[]Store}
// @Failure 500 {object} response.Response
// @Router /stores [get]
func (h *Handler) ListStores(c *gin.Context) {
	stores, err := h.service.ListStores(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Failed to list stores: "+err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	totalCount := len(stores)
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	start := (page - 1) * limit
	if start > totalCount {
		start = totalCount
	}
	end := start + limit
	if end > totalCount {
		end = totalCount
	}

	c.Header("X-Total-Count", strconv.Itoa(totalCount))
	c.Header("X-Page", strconv.Itoa(page))
	c.Header("X-Limit", strconv.Itoa(limit))
	c.Header("X-Total-Pages", strconv.Itoa(totalPages))

	response.Success(c, http.StatusOK, stores[start:end])
}

// GetStore godoc
// @Summary Get store by ID
// @Description Retrieve store details by unique ID
// @Tags Stores
// @Produce json
// @Param id path string true "Store UUID"
// @Success 200 {object} response.Response{data=Store}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /stores/{id} [get]
func (h *Handler) GetStore(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid store ID format")
		return
	}

	store, err := h.service.GetStore(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrStoreNotFound) {
			response.NotFound(c, "Store not found")
			return
		}
		response.InternalServerError(c, "Failed to retrieve store: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, store)
}

// GetMyStore godoc
// @Summary Get current merchant store
// @Description Retrieve the store belonging to the currently logged in merchant
// @Tags Stores
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=Store}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /stores/merchant/mine [get]
func (h *Handler) GetMyStore(c *gin.Context) {
	merchantID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	st, err := h.service.GetMerchantStore(c.Request.Context(), merchantID)
	if err != nil {
		if errors.Is(err, ErrStoreNotFound) {
			response.NotFound(c, "No store found for this merchant")
			return
		}
		response.InternalServerError(c, "Failed to fetch merchant store: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, st)
}

// CreateMenuItem godoc
// @Summary Create menu item
// @Description Add a new menu item to merchant's store
// @Tags Stores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Store UUID"
// @Param request body CreateMenuItemRequest true "Menu item details"
// @Success 201 {object} response.Response{data=MenuItem}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /stores/{id}/menu [post]
func (h *Handler) CreateMenuItem(c *gin.Context) {
	merchantID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	storeIDStr := c.Param("id")
	storeID, err := uuid.Parse(storeIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid store ID format")
		return
	}

	var req CreateMenuItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid menu item payload: "+err.Error())
		return
	}

	item, err := h.service.CreateMenuItem(c.Request.Context(), merchantID, storeID, &req)
	if err != nil {
		if errors.Is(err, ErrStoreNotFound) {
			response.NotFound(c, "Store not found")
			return
		}
		if errors.Is(err, ErrUnauthorizedStoreAccess) {
			response.Forbidden(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, http.StatusCreated, item)
}

// ListMenuItems godoc
// @Summary List menu items for store
// @Description Retrieve all menu items for a specific store
// @Tags Stores
// @Produce json
// @Param id path string true "Store UUID"
// @Success 200 {object} response.Response{data=[]MenuItem}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /stores/{id}/menu [get]
func (h *Handler) ListMenuItems(c *gin.Context) {
	storeIDStr := c.Param("id")
	storeID, err := uuid.Parse(storeIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid store ID format")
		return
	}

	items, err := h.service.ListMenuItems(c.Request.Context(), storeID)
	if err != nil {
		if errors.Is(err, ErrStoreNotFound) {
			response.NotFound(c, "Store not found")
			return
		}
		response.InternalServerError(c, "Failed to list menu items: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, items)
}
