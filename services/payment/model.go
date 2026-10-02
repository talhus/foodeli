package payment

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// statuses
const (
	StatusPending  = "PENDING"
	StatusSuccess  = "SUCCESS"
	StatusFailed   = "FAILED"
	StatusRefunded = "REFUNDED"
)

// payment methods & providers
const (
	MethodCreditCard   = "CREDIT_CARD"
	MethodCreditWallet = "WALLET"

	ProviderMock       = "MOCK"
	ProviderTyrFintech = "TYR_FINTECH"
)

// domain errors
var (
	ErrPaymentNotFound           = errors.New("payment not found")
	ErrOrderNotFound             = errors.New("order not found")
	ErrOrderNotPayable           = errors.New("order is not in payable state")
	ErrAmountMismatch            = errors.New("payment amount does not match order total")
	ErrPaymentFailed             = errors.New("payment processing failed")
	ErrInvalidCardDetails        = errors.New("invalid card details")
	ErrUnauthorizedPaymentAccess = errors.New("unauthorized payment access")
)

type Payment struct {
	ID            uuid.UUID `json:"id"`
	OrderID       uuid.UUID `json:"order_id"`
	UserID        uuid.UUID `json:"user_id"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	Provider      string    `json:"provider"`
	PaymentMethod string    `json:"payment_method"`
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	FailureReason *string   `json:"failure_reason"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ProcessPaymentRequest struct {
	OrderID        uuid.UUID `json:"order_id" binding:"required"`
	PaymentMethod  string    `json:"payment_method" binding:"required"`
	Provider       string    `json:"provider"`
	CardNumber     string    `json:"card_number" binding:"required,min=16,max=16"`
	CardHolderName string    `json:"card_holder_name" binding:"required,min=2,max=100"`
	ExpireMonth    string    `json:"expire_month" binding:"required,min=2,max=2"`
	ExpireYear     string    `json:"expire_year" binding:"required,min=2,max=4"`
	CVV            string    `json:"cvv" binding:"required,min=3,max=4"`
}

type CreateCheckoutSessionRequest struct {
	OrderID uuid.UUID `json:"order_id" binding:"required"`
}

type GatewayCheckoutSessionRequest struct {
	OrderID     string `json:"order_id"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	CallbackURL string `json:"callback_url"`
	WebhookURL  string `json:"webhook_url"`
}

type CheckoutSessionResponse struct {
	SessionID   string `json:"session_id"`
	CheckoutURL string `json:"checkout_url"`
	Status      string `json:"status"`
	ExpiresAt   string `json:"expires_at,omitempty"`
}

type TyrFintechWebhookPayload struct {
	Event         string `json:"event" binding:"required"`
	TransactionID string `json:"transaction_id" binding:"required"`
	SessionID     string `json:"session_id"`
	OrderID       string `json:"order_id" binding:"required"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	Status        string `json:"status" binding:"required"`
	Timestamp     int64  `json:"timestamp"`
}


