package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ChargeRequest struct {
	Amount         int64  `json:"amount"`
	Currency       string `json:"currency"`
	Provider       string `json:"provider,omitempty"`
	CardNumber     string `json:"card_number"`
	CardHolderName string `json:"card_holder_name"`
	ExpireMonth    string `json:"expire_month"`
	ExpireYear     string `json:"expire_year"`
	CVV            string `json:"cvv"`
}
type ChargeResponse struct {
	TransactionID string `json:"transaction_id"`
	Success       bool   `json:"success"`
	FailureReason string `json:"failure_reason,omitempty"`
}

type PaymentGateway interface {
	Charge(ctx context.Context, req *ChargeRequest) (*ChargeResponse, error)
	CreateCheckoutSession(ctx context.Context, req *GatewayCheckoutSessionRequest) (*CheckoutSessionResponse, error)
}

type MockGateway struct{}

func NewMockGateway() *MockGateway {
	return &MockGateway{}
}

func (m *MockGateway) Charge(ctx context.Context, req *ChargeRequest) (*ChargeResponse, error) {
	cleanCard := strings.ReplaceAll(req.CardNumber, " ", "")
	if strings.HasPrefix(cleanCard, "4000") {
		return &ChargeResponse{
			Success:       false,
			FailureReason: "insufficent funds on card",
		}, nil
	}

	if req.CVV == "000" {
		return &ChargeResponse{
			Success:       false,
			FailureReason: "invalid CVV code",
		}, nil
	}
	return &ChargeResponse{
		Success:       true,
		TransactionID: fmt.Sprintf("txn_mock_%s", uuid.New().String()[:12]),
	}, nil
}

func (m *MockGateway) CreateCheckoutSession(ctx context.Context, req *GatewayCheckoutSessionRequest) (*CheckoutSessionResponse, error) {
	sessionID := fmt.Sprintf("cs_mock_%s", uuid.New().String()[:12])
	checkoutBase := req.CallbackURL
	if checkoutBase == "" {
		checkoutBase = "/orders/callback"
	}
	separator := "?"
	if strings.Contains(checkoutBase, "?") {
		separator = "&"
	}
	return &CheckoutSessionResponse{
		SessionID:   sessionID,
		CheckoutURL: fmt.Sprintf("%s%sstatus=SUCCEEDED&order_id=%s&session_id=%s", checkoutBase, separator, req.OrderID, sessionID),
		Status:      "PENDING",
		ExpiresAt:   time.Now().Add(30 * time.Minute).Format(time.RFC3339),
	}, nil
}

//=======================
//	TYR-FINTECH B2B GATEWAY
//=======================

type TyrFintechGateway struct {
	baseURL     string
	apiKey      string
	callbackURL string
	webhookURL  string
	httpClient  *http.Client
}

func NewTyrFintechGateway(baseURL string, apiKey string, callbackURL string, webhookURL string) *TyrFintechGateway {
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}

	return &TyrFintechGateway{
		baseURL:     strings.TrimRight(baseURL, "/"),
		apiKey:      apiKey,
		callbackURL: callbackURL,
		webhookURL:  webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type tyrFintechPayload struct {
	Amount         int64  `json:"amount"`
	Currency       string `json:"currency"`
	CardNumber     string `json:"card_number"`
	CardHolderName string `json:"card_holder_name"`
	ExpireMonth    string `json:"expire_month"`
	ExpireYear     string `json:"expire_year"`
	CVV            string `json:"cvv"`
}

type tyrFintechResponse struct {
	Success       bool   `json:"success"`
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	FailureReason string `json:"failure_reason,omitempty"`
}

func (g *TyrFintechGateway) Charge(ctx context.Context, req *ChargeRequest) (*ChargeResponse, error) {
	endPoint := fmt.Sprintf("%s/api/v1/charges", g.baseURL)
	currency := req.Currency
	if currency == "" {
		currency = "EUR"
	}

	payload := tyrFintechPayload{
		Amount:         req.Amount,
		Currency:       currency,
		CardNumber:     req.CardNumber,
		CardHolderName: req.CardHolderName,
		ExpireMonth:    req.ExpireMonth,
		ExpireYear:     req.ExpireYear,
		CVV:            req.CVV,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tyr-fintech payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endPoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to build TyrFintech HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if g.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)
	}
	// Generate unique idempotency key per charge attempt to prevent double charging on retries
	httpReq.Header.Set("Idempotency-Key", uuid.New().String())

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to reach TyrFintech service at %s: %w", g.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("tyrfintech authorization failed: invalid or missing api key")
	}

	var tyrResp tyrFintechResponse
	if err := json.NewDecoder(resp.Body).Decode(&tyrResp); err != nil {
		return nil, fmt.Errorf("failed to decode tyrfintech response: %w", err)
	}

	// HTTP 200/201: Success or explicit decline reported in 200 body
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		return &ChargeResponse{
			Success:       tyrResp.Success,
			TransactionID: tyrResp.TransactionID,
			FailureReason: tyrResp.FailureReason,
		}, nil
	}

	// HTTP 400/402: Card declined, expired, or insufficient funds
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusPaymentRequired {
		failureReason := tyrResp.FailureReason
		if failureReason == "" {
			failureReason = "Payment declined by issuing bank"
		}
		return &ChargeResponse{
			Success:       false,
			TransactionID: tyrResp.TransactionID,
			FailureReason: failureReason,
		}, nil
	}

	return nil, fmt.Errorf("unexpected status from tyrfintech (%d): %s", resp.StatusCode, tyrResp.FailureReason)
}

func (g *TyrFintechGateway) CreateCheckoutSession(ctx context.Context, req *GatewayCheckoutSessionRequest) (*CheckoutSessionResponse, error) {
	endPoint := fmt.Sprintf("%s/api/v1/checkout/sessions", g.baseURL)
	callbackURL := req.CallbackURL
	if callbackURL == "" {
		callbackURL = g.callbackURL
	}
	webhookURL := req.WebhookURL
	if webhookURL == "" {
		webhookURL = g.webhookURL
	}
	currency := req.Currency
	if currency == "" {
		currency = "EUR"
	}

	payload := map[string]any{
		"order_id":     req.OrderID,
		"amount":       req.Amount,
		"currency":     currency,
		"callback_url": callbackURL,
		"webhook_url":  webhookURL,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal checkout session payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endPoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to build TyrFintech checkout session request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if g.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)
	}

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to reach TyrFintech service at %s: %w", g.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("tyrfintech authorization failed: invalid or missing api key")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errMap map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errMap)
		return nil, fmt.Errorf("tyrfintech session creation failed with status %d: %v", resp.StatusCode, errMap)
	}

	var wrapper struct {
		SessionID   string `json:"session_id"`
		CheckoutURL string `json:"checkout_url"`
		Status      string `json:"status"`
		ExpiresAt   string `json:"expires_at"`
		Data        *struct {
			SessionID   string `json:"session_id"`
			CheckoutURL string `json:"checkout_url"`
			Status      string `json:"status"`
			ExpiresAt   string `json:"expires_at"`
		} `json:"data,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("failed to decode tyrfintech session response: %w", err)
	}

	res := &CheckoutSessionResponse{
		SessionID:   wrapper.SessionID,
		CheckoutURL: wrapper.CheckoutURL,
		Status:      wrapper.Status,
		ExpiresAt:   wrapper.ExpiresAt,
	}
	if res.SessionID == "" && wrapper.Data != nil {
		res.SessionID = wrapper.Data.SessionID
		res.CheckoutURL = wrapper.Data.CheckoutURL
		res.Status = wrapper.Data.Status
		res.ExpiresAt = wrapper.Data.ExpiresAt
	}

	return res, nil
}

// ======================================================================
// 3. MULTI-GATEWAY DISPATCHER
// ======================================================================

type MultiGateway struct {
	gateways  map[string]PaymentGateway
	defaultGW PaymentGateway
}

func NewMultiGateway(gateways map[string]PaymentGateway, defaultGW PaymentGateway) *MultiGateway {
	return &MultiGateway{
		gateways:  gateways,
		defaultGW: defaultGW,
	}
}

func (m *MultiGateway) Charge(ctx context.Context, req *ChargeRequest) (*ChargeResponse, error) {
	gw, ok := m.gateways[req.Provider]
	if !ok {
		gw = m.defaultGW
	}

	res, err := gw.Charge(ctx, req)
	if err != nil && req.Provider == ProviderTyrFintech {
		log.Printf("[MultiGateway] TyrFintech charge failed (%v), falling back to MockGateway", err)
		return m.defaultGW.Charge(ctx, req)
	}
	return res, err
}

func (m *MultiGateway) CreateCheckoutSession(ctx context.Context, req *GatewayCheckoutSessionRequest) (*CheckoutSessionResponse, error) {
	if gw, ok := m.gateways[ProviderTyrFintech]; ok {
		sess, err := gw.CreateCheckoutSession(ctx, req)
		if err == nil {
			return sess, nil
		}
		log.Printf("[MultiGateway] TyrFintech checkout session creation failed (%v), falling back to MockGateway", err)
	}
	return m.defaultGW.CreateCheckoutSession(ctx, req)
}

