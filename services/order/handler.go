package order

import (
	"errors"
	"foodeli/platform/middleware"
	"foodeli/platform/response"
	"foodeli/services/store"
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup, authMW gin.HandlerFunc) {
	orders := r.Group("/orders")
	orders.Use(authMW)

	// Customer endpoints
	orders.POST("", middleware.RequireRole("customer"), h.CreateOrder)
	orders.GET("/customer/mine", middleware.RequireRole("customer"), h.ListCustomerOrders)

	// Merchant endpoints
	orders.GET("/merchant/mine", middleware.RequireRole("merchant"), h.ListMerchantOrders)
	orders.PATCH("/:id/status", middleware.RequireRole("merchant"), h.UpdateOrderStatus)

	// Single order detail (Customer, Merchant, Admin)
	orders.GET("/:id", h.GetOrderById)
}

// =============================
// METHODS
// =============================

// CreateOrder godoc
// @Summary Create an order directly
// @Description Place an order directly with items and address
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Idempotency-Key header string false "Idempotency key UUID"
// @Param request body CreateOrderRequest true "Create order payload"
// @Success 201 {object} response.Response{data=Order}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	customerID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	// Read standard Stripe-style Idempotency-Key header
	idempotencyKey := c.GetHeader("Idempotency-Key")

	order, err := h.service.CreateOrder(c, customerID, &req, idempotencyKey)
	if err != nil {
		if errors.Is(err, ErrStoreNotFound) {
			response.NotFound(c, "Store not found")
			return
		}
		if errors.Is(err, ErrProductNotFound) {
			response.NotFound(c, "One or more products were not found in this store")
			return
		}
		if errors.Is(err, ErrInsufficientStock) {
			response.BadRequest(c, "Insufficient stock for one or more requested products")
			return
		}
		if errors.Is(err, ErrEmptyOrder) {
			response.BadRequest(c, "Order must contain at least one item")
			return
		}
		response.InternalServerError(c, "Failed to create order: "+err.Error())
		return
	}

	response.Success(c, 201, order)
}

// GetOrderById godoc
// @Summary Get order by ID
// @Description Retrieve order details by ID for owner customer or store merchant
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order UUID"
// @Success 200 {object} response.Response{data=Order}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/{id} [get]
func (h *Handler) GetOrderById(c *gin.Context) {
	requesterID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	role, ok := middleware.GetUserRole(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid order ID format")
		return
	}

	order, err := h.service.GetOrderById(c, requesterID, role, orderID)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			response.NotFound(c, "Order not found")
			return
		}
		if errors.Is(err, ErrUnauthorizedOrderAccess) {
			response.Forbidden(c, "You do not have permission to view this order")
			return
		}
		response.InternalServerError(c, "Failed to retrieve order: "+err.Error())
		return
	}

	response.Success(c, 200, order)
}

// ListCustomerOrders godoc
// @Summary List customer orders
// @Description Retrieve order history for the authenticated customer
// @Tags Orders
// @Produce json
// @Security BearerAuth
// paginateOrders clamps page/limit, sets pagination headers, and slices orders.
func paginateOrders(c *gin.Context, orders []Order) []Order {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	totalCount := len(orders)
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

	return orders[start:end]
}

// @Success 200 {object} response.Response{data=[]Order}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/customer/mine [get]
func (h *Handler) ListCustomerOrders(c *gin.Context) {
	customerID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	orders, err := h.service.ListCustomerOrders(c, customerID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve customer orders: "+err.Error())
		return
	}

	paged := paginateOrders(c, orders)
	response.Success(c, 200, paged)
}

// ListMerchantOrders godoc
// @Summary List merchant orders
// @Description Retrieve all incoming orders for the merchant's store
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]Order}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/merchant/mine [get]
func (h *Handler) ListMerchantOrders(c *gin.Context) {
	merchantID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	orders, err := h.service.ListMerchantOrders(c, merchantID)
	if err != nil {
		if errors.Is(err, store.ErrStoreNotFound) {
			response.NotFound(c, "Store not found for this merchant")
			return
		}
		response.InternalServerError(c, "Failed to retrieve merchant orders: "+err.Error())
		return
	}

	paged := paginateOrders(c, orders)
	response.Success(c, 200, paged)
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Description Transition order status (PENDING -> CONFIRMED -> PREPARING -> DELIVERED or CANCELLED)
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order UUID"
// @Param request body UpdateOrderStatusRequest true "Order status payload"
// @Success 200 {object} response.Response{data=Order}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/{id}/status [patch]
func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	merchantID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid order ID format")
		return
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload")
		return
	}

	order, err := h.service.UpdateOrderStatus(c, merchantID, orderID, req.Status)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			response.NotFound(c, "Order not found")
			return
		}
		if errors.Is(err, store.ErrStoreNotFound) {
			response.NotFound(c, "Store not found for this merchant")
			return
		}
		if errors.Is(err, ErrUnauthorizedOrderAccess) {
			response.Forbidden(c, "You do not have permission to manage this order")
			return
		}
		if errors.Is(err, ErrInvalidOrderStatus) {
			response.BadRequest(c, "Invalid order status transition")
			return
		}
		response.InternalServerError(c, "Failed to update order status: "+err.Error())
		return
	}

	response.Success(c, 200, order)
}
