package config

import "testing"

func TestLoadConfigFromEnvironment(t *testing.T) {
	t.Setenv("SHATRANJ_HTTP_ADDR", "127.0.0.1:9090")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost/db")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("SHATRANJ_CORS_ORIGINS", "https://example.com, https://app.example.com")
	t.Setenv("SHATRANJ_SHUTDOWN_TIMEOUT", "15s")
	t.Setenv("SHATRANJ_PREMIUM_PRICE_PAISE", "12500")
	t.Setenv("RAZORPAY_KEY_ID", "")
	t.Setenv("RAZORPAY_KEY_SECRET", "")
	t.Setenv("RAZORPAY_WEBHOOK_SECRET", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPAddr != "127.0.0.1:9090" || cfg.PostgresDSN == "" || cfg.RedisURL == "" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if len(cfg.CORSOrigins) != 2 || cfg.CORSOrigins[1] != "https://app.example.com" {
		t.Fatalf("CORSOrigins = %#v", cfg.CORSOrigins)
	}
	if cfg.ShutdownTimeout.String() != "15s" {
		t.Fatalf("ShutdownTimeout = %s, want 15s", cfg.ShutdownTimeout)
	}
	if cfg.PremiumPricePaise != 12500 {
		t.Fatalf("PremiumPricePaise = %d, want 12500", cfg.PremiumPricePaise)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	t.Setenv("SHATRANJ_HTTP_ADDR", "localhost:not-a-port")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted a non-numeric port")
	}

	t.Setenv("SHATRANJ_HTTP_ADDR", ":8080")
	t.Setenv("SHATRANJ_SHUTDOWN_TIMEOUT", "0s")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted a non-positive shutdown timeout")
	}

	t.Setenv("SHATRANJ_SHUTDOWN_TIMEOUT", "")
	t.Setenv("SHATRANJ_PREMIUM_PRICE_PAISE", "0")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted a non-positive premium price")
	}
	t.Setenv("SHATRANJ_PREMIUM_PRICE_PAISE", "")
	t.Setenv("RAZORPAY_KEY_ID", "key_id")
	t.Setenv("RAZORPAY_KEY_SECRET", "")
	t.Setenv("RAZORPAY_WEBHOOK_SECRET", "")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted an incomplete Razorpay configuration")
	}
}
