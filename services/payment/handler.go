package payment

import (
	"errors"
	"net/http"

	"foodeli/platform/middleware"
	"foodeli/platform/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup, authMW gin.HandlerFunc, rateLimiters ...gin.HandlerFunc) {
	payments := r.Group("/payments")
	payments.Use(authMW)

	// Customer checkout payment endpoint
	payments.POST("/process", middleware.RequireRole("customer"), h.ProcessPayment)

	// Customer hosted checkout session endpoint ("Pay with TyrFintech Wallet")
	payments.POST("/checkout-session", middleware.RequireRole("customer"), h.CreateCheckoutSession)

	// Order payment receipt inquiry
	payments.GET("/order/:orderId", h.GetPaymentByOrderID)

	// Webhooks (public server-to-server endpoints from payment partners)
	webhooks := r.Group("/webhooks")
	if len(rateLimiters) > 0 && rateLimiters[0] != nil {
		webhooks.POST("/tyrfintech", rateLimiters[0], h.HandleTyrFintechWebhook)
	} else {
		webhooks.POST("/tyrfintech", h.HandleTyrFintechWebhook)
	}
}

// ProcessPayment godoc
// @Summary Process payment (Direct Card)
// @Description Process direct credit card payment for an order via TyrFintech
// @Tags Payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ProcessPaymentRequest true "Payment payload"
// @Success 200 {object} response.Response{data=Payment}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 402 {object} response.Response{data=Payment} "Payment processing declined"
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /payments/process [post]
func (h *Handler) ProcessPayment(c *gin.Context) {
	customerID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	payment, err := h.service.ProcessPayment(c, customerID, &req)
	if err != nil {
		if errors.Is(err, ErrPaymentFailed) {
			// HTTP 402 Payment Required: Payment was declined by gateway
			c.JSON(http.StatusPaymentRequired, gin.H{
				"status":  "error",
				"message": "Payment processing declined: " + *payment.FailureReason,
				"data":    payment,
			})
			return
		}
		if errors.Is(err, ErrOrderNotFound) {
			response.NotFound(c, "Order not found")
			return
		}
		if errors.Is(err, ErrOrderNotPayable) {
			response.BadRequest(c, "Order is not payable (must be PENDING)")
			return
		}
		if errors.Is(err, ErrUnauthorizedPaymentAccess) {
			response.Forbidden(c, "You do not have permission to pay for this order")
			return
		}
		response.InternalServerError(c, "Failed to process payment: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, payment)
}

// GetPaymentByOrderID godoc
// @Summary Get payment by order ID
// @Description Retrieve payment receipt details for an order
// @Tags Payments
// @Produce json
// @Security BearerAuth
// @Param orderId path string true "Order UUID"
// @Success 200 {object} response.Response{data=Payment}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /payments/order/{orderId} [get]
func (h *Handler) GetPaymentByOrderID(c *gin.Context) {
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

	orderID, err := uuid.Parse(c.Param("orderId"))
	if err != nil {
		response.BadRequest(c, "Invalid order ID format")
		return
	}

	payment, err := h.service.GetPaymentByOrderId(c, requesterID, role, orderID)
	if err != nil {
		if errors.Is(err, ErrPaymentNotFound) {
			response.NotFound(c, "Payment record not found for this order")
			return
		}
		if errors.Is(err, ErrUnauthorizedPaymentAccess) {
			response.Forbidden(c, "You do not have permission to view this payment")
			return
		}
		response.InternalServerError(c, "Failed to retrieve payment: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, payment)
}

// CreateCheckoutSession godoc
// @Summary Create hosted checkout session
// @Description Initiate a TyrFintech hosted checkout session for wallet or external payment
// @Tags Payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateCheckoutSessionRequest true "Checkout session payload"
// @Success 200 {object} response.Response{data=CheckoutSessionResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /payments/checkout-session [post]
func (h *Handler) CreateCheckoutSession(c *gin.Context) {
	customerID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req CreateCheckoutSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	session, err := h.service.CreateCheckoutSession(c, customerID, &req)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			response.NotFound(c, "Order not found")
			return
		}
		if errors.Is(err, ErrOrderNotPayable) {
			response.BadRequest(c, "Order is not payable (must be PENDING)")
			return
		}
		if errors.Is(err, ErrUnauthorizedPaymentAccess) {
			response.Forbidden(c, "You do not have permission to pay for this order")
			return
		}
		response.InternalServerError(c, "Failed to create checkout session: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, session)
}

// HandleTyrFintechWebhook godoc
// @Summary TyrFintech payment webhook
// @Description Receive and idempotently process asynchronous payment events from TyrFintech
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body TyrFintechWebhookPayload true "Webhook payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /webhooks/tyrfintech [post]
func (h *Handler) HandleTyrFintechWebhook(c *gin.Context) {
	var payload TyrFintechWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.BadRequest(c, "Invalid webhook payload: "+err.Error())
		return
	}

	if err := h.service.HandleTyrFintechWebhook(c.Request.Context(), &payload); err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			response.NotFound(c, "Order not found")
			return
		}
		response.InternalServerError(c, "Failed to process webhook: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"received": true,
	})
}


