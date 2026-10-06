package payments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const razorpayOrdersURL = "https://api.razorpay.com/v1/orders"

type RazorpayClient struct {
	keyID         string
	keySecret     string
	webhookSecret string
	client        *http.Client
}

func NewRazorpayClient(keyID, keySecret, webhookSecret string) *RazorpayClient {
	return &RazorpayClient{
		keyID: keyID, keySecret: keySecret, webhookSecret: webhookSecret,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *RazorpayClient) KeyID() string { return c.keyID }

func (c *RazorpayClient) CreateOrder(ctx context.Context, amount int64, currency, receipt string) (string, error) {
	if c.keyID == "" || c.keySecret == "" {
		return "", ErrProviderUnavailable
	}
	body, err := json.Marshal(struct {
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
		Receipt  string `json:"receipt"`
	}{Amount: amount, Currency: currency, Receipt: receipt})
	if err != nil {
		return "", fmt.Errorf("encode Razorpay order: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, razorpayOrdersURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create Razorpay request: %w", err)
	}
	request.SetBasicAuth(c.keyID, c.keySecret)
	request.Header.Set("Content-Type", "application/json")
	response, err := c.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("create Razorpay order: %w", err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read Razorpay response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("Razorpay order API returned status %d", response.StatusCode)
	}
	var order struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(responseBody, &order); err != nil {
		return "", fmt.Errorf("decode Razorpay order response: %w", err)
	}
	if !strings.HasPrefix(order.ID, "order_") {
		return "", fmt.Errorf("Razorpay returned invalid order ID")
	}
	return order.ID, nil
}

func (c *RazorpayClient) VerifyWebhook(body []byte, signature string) bool {
	return verifySignature(c.webhookSecret, body, signature)
}

var _ Provider = (*RazorpayClient)(nil)
