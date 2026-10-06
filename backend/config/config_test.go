package config

import "testing"

func TestLoadUsesDefaultAddress(t *testing.T) {
	t.Setenv("SHATRANJ_HTTP_ADDR", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned an error: %v", err)
	}
	if cfg.HTTPAddr != defaultHTTPAddr {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, defaultHTTPAddr)
	}
}

func TestLoadUsesConfiguredAddress(t *testing.T) {
	t.Setenv("SHATRANJ_HTTP_ADDR", "127.0.0.1:9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned an error: %v", err)
	}
	if cfg.HTTPAddr != "127.0.0.1:9090" {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, "127.0.0.1:9090")
	}
}

func TestLoadRejectsInvalidAddress(t *testing.T) {
	t.Setenv("SHATRANJ_HTTP_ADDR", "localhost")

	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted an address without a port")
	}
}
