package cart

import (
	"errors"
	"net/http"

	"foodeli/platform/middleware"
	"foodeli/platform/response"
	"foodeli/services/order"

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
	cart := r.Group("/cart")
	cart.Use(authMW)
	cart.Use(middleware.RequireRole("customer"))

	cart.GET("", h.GetCart)
	cart.POST("/items", h.AddItem)
	cart.PUT("/items/:productId", h.UpdateItemQuantity)
	cart.DELETE("/items/:productId", h.RemoveItem)
	cart.DELETE("", h.ClearCart)
	cart.POST("/checkout", h.Checkout)
}

// GetCart godoc
// @Summary Get customer shopping cart
// @Description Retrieve customer's active shopping cart stored in Redis
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=Cart}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cart [get]
func (h *Handler) GetCart(c *gin.Context) {
	customerID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	cart, err := h.service.GetCart(c, customerID)
	if err != nil {
		response.InternalServerError(c, "Failed to retrieve cart: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, cart)
}

// AddItem godoc
// @Summary Add item to cart
// @Description Add a product with quantity to cart (enforces single-store cart unless force=true)
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AddToCartRequest true "Add to cart payload"
// @Success 200 {object} response.Response{data=Cart}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cart/items [post]
func (h *Handler) AddItem(c *gin.Context) {
	customerID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	cart, err := h.service.AddItem(c, customerID, &req)
	if err != nil {
		if errors.Is(err, ErrInvalidQuantity) {
			response.BadRequest(c, "Quantity must be greater than zero")
			return
		}
		if errors.Is(err, ErrProductNotFound) {
			response.NotFound(c, "Product or store not found")
			return
		}
		if errors.Is(err, ErrInsufficientStock) {
			response.BadRequest(c, "Insufficient stock for this product")
			return
		}
		if errors.Is(err, ErrCrossStoreConflict) {
			response.Conflict(c, "Cart contains items from another store. Set force=true to clear and replace your cart.")
			return
		}
		response.InternalServerError(c, "Failed to add item to cart: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, cart)
}

// UpdateItemQuantity godoc
// @Summary Update cart item quantity
// @Description Modify the quantity of a product in the cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param productId path string true "Product UUID"
// @Param request body UpdateCartItemRequest true "Update quantity payload"
// @Success 200 {object} response.Response{data=Cart}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cart/items/{productId} [put]
func (h *Handler) UpdateItemQuantity(c *gin.Context) {
	customerID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		response.BadRequest(c, "Invalid product ID format")
		return
	}

	var req UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	cart, err := h.service.UpdateItemQuantity(c, customerID, productID, req.Quantity)
	if err != nil {
		if errors.Is(err, ErrCartNotFound) || errors.Is(err, ErrItemNotFoundInCart) {
			response.NotFound(c, "Item not found in cart")
			return
		}
		if errors.Is(err, ErrProductNotFound) {
			response.NotFound(c, "Product not found")
			return
		}
		if errors.Is(err, ErrInsufficientStock) {
			response.BadRequest(c, "Insufficient stock for this product")
			return
		}
		response.InternalServerError(c, "Failed to update item quantity: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, cart)
}

// RemoveItem godoc
// @Summary Remove item from cart
// @Description Remove a specific product from customer's cart
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Param productId path string true "Product UUID"
// @Success 200 {object} response.Response{data=Cart}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cart/items/{productId} [delete]
func (h *Handler) RemoveItem(c *gin.Context) {
	customerID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		response.BadRequest(c, "Invalid product ID format")
		return
	}

	cart, err := h.service.RemoveItem(c, customerID, productID)
	if err != nil {
		if errors.Is(err, ErrCartNotFound) || errors.Is(err, ErrItemNotFoundInCart) {
			response.NotFound(c, "Item not found in cart")
			return
		}
		response.InternalServerError(c, "Failed to remove item from cart: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, cart)
}

// ClearCart godoc
// @Summary Clear active cart
// @Description Empty all items in customer's cart
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cart [delete]
func (h *Handler) ClearCart(c *gin.Context) {
	customerID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	if err := h.service.ClearCart(c, customerID); err != nil {
		response.InternalServerError(c, "Failed to clear cart: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message": "Cart cleared successfully",
	})
}

// Checkout godoc
// @Summary Checkout cart to create order
// @Description Convert active cart into a confirmed pending order and atomic stock reservation
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Idempotency-Key header string false "Idempotency key UUID"
// @Param request body CheckoutRequest true "Checkout payload"
// @Success 201 {object} response.Response{data=order.Order}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cart/checkout [post]
func (h *Handler) Checkout(c *gin.Context) {
	customerID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	idempotencyKey := c.GetHeader("Idempotency-Key")

	createdOrder, err := h.service.Checkout(c, customerID, &req, idempotencyKey)
	if err != nil {
		if errors.Is(err, ErrCartEmpty) || errors.Is(err, ErrCartNotFound) {
			response.BadRequest(c, "Cannot checkout an empty cart")
			return
		}
		if errors.Is(err, order.ErrStoreNotFound) {
			response.NotFound(c, "Store not found")
			return
		}
		if errors.Is(err, order.ErrInsufficientStock) {
			response.BadRequest(c, "One or more products have insufficient stock")
			return
		}
		response.InternalServerError(c, "Checkout failed: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, createdOrder)
}
