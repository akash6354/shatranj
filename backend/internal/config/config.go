package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultHTTPAddr = ":8080"

// Config contains process-level settings shared by the server and worker.
type Config struct {
	HTTPAddr    string
	PostgresDSN string
	// RedisURL enables Redis-backed rate limiting and presence. It is optional:
	// when empty, unparseable, or unreachable, the process falls back to
	// in-memory stores instead of failing to start.
	RedisURL              string
	JWTSecret             string
	CORSOrigins           []string
	ShutdownTimeout       time.Duration
	PremiumPricePaise     int64
	RazorpayKeyID         string
	RazorpayKeySecret     string
	RazorpayWebhookSecret string
}

// Load reads configuration from the environment and validates its values.
func Load() (Config, error) {
	addr := strings.TrimSpace(os.Getenv("SHATRANJ_HTTP_ADDR"))
	if addr == "" {
		addr = defaultHTTPAddr
	}
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return Config{}, fmt.Errorf("invalid SHATRANJ_HTTP_ADDR %q: %w", addr, err)
	}
	if _, err := strconv.ParseUint(port, 10, 16); err != nil {
		return Config{}, fmt.Errorf("invalid port in SHATRANJ_HTTP_ADDR %q: %w", addr, err)
	}

	timeout := 10 * time.Second
	premiumPrice := int64(49_900)
	if raw := strings.TrimSpace(os.Getenv("SHATRANJ_PREMIUM_PRICE_PAISE")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 || parsed > 100_000_000 {
			return Config{}, fmt.Errorf("invalid SHATRANJ_PREMIUM_PRICE_PAISE %q: must be between 1 and 100000000", raw)
		}
		premiumPrice = parsed
	}
	if raw := strings.TrimSpace(os.Getenv("SHATRANJ_SHUTDOWN_TIMEOUT")); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil || parsed <= 0 {
			return Config{}, fmt.Errorf("invalid SHATRANJ_SHUTDOWN_TIMEOUT %q: must be a positive duration", raw)
		}
		timeout = parsed
	}
	keyID := strings.TrimSpace(os.Getenv("RAZORPAY_KEY_ID"))
	keySecret := os.Getenv("RAZORPAY_KEY_SECRET")
	webhookSecret := os.Getenv("RAZORPAY_WEBHOOK_SECRET")
	razorpayConfigured := keyID != "" || keySecret != "" || webhookSecret != ""
	if razorpayConfigured && (keyID == "" || keySecret == "" || webhookSecret == "") {
		return Config{}, fmt.Errorf("RAZORPAY_KEY_ID, RAZORPAY_KEY_SECRET, and RAZORPAY_WEBHOOK_SECRET must be configured together")
	}

	return Config{
		HTTPAddr:              addr,
		PostgresDSN:           strings.TrimSpace(os.Getenv("DATABASE_URL")),
		RedisURL:              strings.TrimSpace(os.Getenv("REDIS_URL")),
		JWTSecret:             os.Getenv("SHATRANJ_JWT_SECRET"),
		CORSOrigins:           splitList(os.Getenv("SHATRANJ_CORS_ORIGINS")),
		ShutdownTimeout:       timeout,
		PremiumPricePaise:     premiumPrice,
		RazorpayKeyID:         keyID,
		RazorpayKeySecret:     keySecret,
		RazorpayWebhookSecret: webhookSecret,
	}, nil
}

func splitList(value string) []string {
	var values []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			values = append(values, item)
		}
	}
	return values
}
