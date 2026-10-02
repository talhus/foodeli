package payment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"foodeli/platform/events"
	"foodeli/platform/outbox"
	"foodeli/services/order"

	"github.com/google/uuid"
)

type OrderManager interface {
	GetOrderById(ctx context.Context, requesterID uuid.UUID, role string, orderID uuid.UUID) (*order.Order, error)
	MarkOrderPaid(ctx context.Context, orderID uuid.UUID) error
}

type Service interface {
	ProcessPayment(ctx context.Context, userID uuid.UUID, req *ProcessPaymentRequest) (*Payment, error)
	GetPaymentByOrderId(ctx context.Context, requesterID uuid.UUID, role string, orderID uuid.UUID) (*Payment, error)
	CreateCheckoutSession(ctx context.Context, userID uuid.UUID, req *CreateCheckoutSessionRequest) (*CheckoutSessionResponse, error)
	HandleTyrFintechWebhook(ctx context.Context, payload *TyrFintechWebhookPayload) error
}

type PaymentService struct {
	repo           Repository
	orderManager   OrderManager
	gateway        PaymentGateway
	eventPublisher EventPublisher
	outboxRepo     outbox.Repository
	callbackURL    string
	webhookURL     string
}

type EventPublisher interface {
	PublishJSON(ctx context.Context, routingKey string, payload any) error
}

func NewService(repo Repository, orderManager OrderManager, gateway PaymentGateway, eventPublisher EventPublisher, outboxRepo outbox.Repository, callbackURL, webhookURL string) *PaymentService {
	return &PaymentService{
		repo:           repo,
		orderManager:   orderManager,
		gateway:        gateway,
		eventPublisher: eventPublisher,
		outboxRepo:     outboxRepo,
		callbackURL:    callbackURL,
		webhookURL:     webhookURL,
	}
}

// ======================================================================
// METHODS
// ======================================================================
func (s *PaymentService) ProcessPayment(ctx context.Context, userID uuid.UUID, req *ProcessPaymentRequest) (*Payment, error) {
	//get order
	ord, err := s.orderManager.GetOrderById(ctx, userID, "customer", req.OrderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}
	//is user owner?
	if ord.UserID != userID {
		return nil, ErrUnauthorizedPaymentAccess
	}

	//is order in valid state to be paid?
	if ord.Status != order.StatusPending {
		return nil, ErrOrderNotPayable
	}

	//is order already paid?
	existing, err := s.repo.GetByOrderId(ctx, req.OrderID)
	if err == nil && existing != nil && existing.Status == StatusSuccess {
		return existing, nil
	}

	//charge
	provider := req.Provider
	if provider == "" {
		provider = ProviderMock
	}
	chargeReq := &ChargeRequest{
		Amount:         ord.TotalAmount,
		Currency:       "EUR",
		Provider:       provider,
		CardNumber:     req.CardNumber,
		CardHolderName: req.CardHolderName,
		ExpireMonth:    req.ExpireMonth,
		ExpireYear:     req.ExpireYear,
		CVV:            req.CVV,
	}

	chargeRes, err := s.gateway.Charge(ctx, chargeReq)
	if err != nil {
		return nil, err
	}

	//ledger and confirm order
	paymentStatus := StatusSuccess
	var failureReason *string
	if !chargeRes.Success {
		paymentStatus = StatusFailed
		failureReason = &chargeRes.FailureReason
	}

	payment := &Payment{
		ID:            uuid.New(),
		OrderID:       ord.ID,
		UserID:        userID,
		Amount:        ord.TotalAmount,
		Currency:      "EUR",
		PaymentMethod: req.PaymentMethod,
		Provider:      provider,
		TransactionID: chargeRes.TransactionID,
		Status:        paymentStatus,
		FailureReason: failureReason,
	}

	if err := s.repo.CreatePayment(ctx, payment); err != nil {
		return nil, err
	}

	//confirmation
	if paymentStatus == StatusSuccess {
		if err := s.orderManager.MarkOrderPaid(ctx, ord.ID); err != nil {
			return nil, err
		}

		eventPayload := events.PaymentSucceededPayload{
			PaymentID:     payment.ID,
			OrderID:       payment.OrderID,
			UserID:        payment.UserID,
			Amount:        payment.Amount,
			PaymentMethod: payment.PaymentMethod,
			TransactionID: payment.TransactionID,
			OccurredAt:    payment.CreatedAt,
		}

		// Atomically/durably persist event into transactional outbox
		if s.outboxRepo != nil {
			_ = s.outboxRepo.Create(ctx, outbox.AggregatePayment, payment.ID, events.PaymentSucceededRoutingKey, eventPayload)
		}

		// Direct event publisher (fast path if available)
		if s.eventPublisher != nil {
			_ = s.eventPublisher.PublishJSON(ctx, events.PaymentSucceededRoutingKey, &eventPayload)
		}
		return payment, nil
	}
	return payment, ErrPaymentFailed
}

func (s *PaymentService) GetPaymentByOrderId(ctx context.Context, requesterID uuid.UUID, role string, orderID uuid.UUID) (*Payment, error) {
	payment, err := s.repo.GetByOrderId(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if role != "admin" && payment.UserID != requesterID {
		return nil, ErrUnauthorizedPaymentAccess
	}
	return payment, nil
}

func (s *PaymentService) CreateCheckoutSession(ctx context.Context, userID uuid.UUID, req *CreateCheckoutSessionRequest) (*CheckoutSessionResponse, error) {
	ord, err := s.orderManager.GetOrderById(ctx, userID, "customer", req.OrderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	if ord.UserID != userID {
		return nil, ErrUnauthorizedPaymentAccess
	}

	if ord.Status != order.StatusPending {
		return nil, ErrOrderNotPayable
	}

	existing, err := s.repo.GetByOrderId(ctx, req.OrderID)
	if err == nil && existing != nil && existing.Status == StatusSuccess {
		return nil, errors.New("order is already paid")
	}

	gwReq := &GatewayCheckoutSessionRequest{
		OrderID:     ord.ID.String(),
		Amount:      ord.TotalAmount,
		Currency:    "EUR",
		CallbackURL: s.callbackURL,
		WebhookURL:  s.webhookURL,
	}

	session, err := s.gateway.CreateCheckoutSession(ctx, gwReq)
	if err != nil {
		return nil, err
	}

	// Pre-seed pending payment record if not already existing
	if existing == nil {
		pendingPayment := &Payment{
			ID:            uuid.New(),
			OrderID:       ord.ID,
			UserID:        userID,
			Amount:        ord.TotalAmount,
			Currency:      "EUR",
			PaymentMethod: MethodCreditWallet,
			Provider:      ProviderTyrFintech,
			TransactionID: session.SessionID,
			Status:        StatusPending,
		}
		_ = s.repo.CreatePayment(ctx, pendingPayment)
	}

	return session, nil
}

func (s *PaymentService) HandleTyrFintechWebhook(ctx context.Context, payload *TyrFintechWebhookPayload) error {
	if payload.Event != "checkout.session.paid" || payload.Status != "SUCCEEDED" {
		// Non-settlement events are safely ignored
		return nil
	}

	orderUUID, err := uuid.Parse(payload.OrderID)
	if err != nil {
		return fmt.Errorf("invalid order_id in webhook payload: %w", err)
	}

	// 1. Fetch order using admin privilege
	ord, err := s.orderManager.GetOrderById(ctx, uuid.Nil, "admin", orderUUID)
	if err != nil {
		return ErrOrderNotFound
	}

	// 2. Idempotency Check: if order is already confirmed / preparing / delivered, skip
	existing, err := s.repo.GetByOrderId(ctx, orderUUID)
	if err == nil && existing != nil && existing.Status == StatusSuccess {
		return nil
	}

	// 3. Mark Order as Paid (transitions PENDING -> CONFIRMED)
	if ord.Status == order.StatusPending {
		if err := s.orderManager.MarkOrderPaid(ctx, orderUUID); err != nil {
			return fmt.Errorf("failed to transition order to paid: %w", err)
		}
	}

	// 4. Record or update payment record
	amount := payload.Amount
	if amount <= 0 {
		amount = ord.TotalAmount
	}
	currency := payload.Currency
	if currency == "" {
		currency = "EUR"
	}

	var paymentRecord *Payment
	if existing != nil {
		if err := s.repo.UpdateTransactionAndStatus(ctx, existing.ID, payload.TransactionID, StatusSuccess); err != nil {
			return fmt.Errorf("failed to update payment record: %w", err)
		}
		existing.TransactionID = payload.TransactionID
		existing.Status = StatusSuccess
		paymentRecord = existing
	} else {
		paymentRecord = &Payment{
			ID:            uuid.New(),
			OrderID:       ord.ID,
			UserID:        ord.UserID,
			Amount:        amount,
			Currency:      currency,
			PaymentMethod: MethodCreditWallet,
			Provider:      ProviderTyrFintech,
			TransactionID: payload.TransactionID,
			Status:        StatusSuccess,
		}
		if err := s.repo.CreatePayment(ctx, paymentRecord); err != nil {
			return fmt.Errorf("failed to persist payment record: %w", err)
		}
	}

	eventPayload := events.PaymentSucceededPayload{
		PaymentID:     paymentRecord.ID,
		OrderID:       paymentRecord.OrderID,
		UserID:        paymentRecord.UserID,
		Amount:        paymentRecord.Amount,
		PaymentMethod: paymentRecord.PaymentMethod,
		TransactionID: paymentRecord.TransactionID,
		OccurredAt:    time.Now().UTC(),
	}

	// Atomically/durably persist event into transactional outbox
	if s.outboxRepo != nil {
		_ = s.outboxRepo.Create(ctx, outbox.AggregatePayment, paymentRecord.ID, events.PaymentSucceededRoutingKey, eventPayload)
	}

	// 5. Direct event publish to RabbitMQ (fast path if available)
	if s.eventPublisher != nil {
		_ = s.eventPublisher.PublishJSON(ctx, events.PaymentSucceededRoutingKey, &eventPayload)
	}

	return nil
}


