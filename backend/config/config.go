package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const defaultHTTPAddr = ":8080"

// Config contains process-level configuration loaded from the environment.
type Config struct {
	HTTPAddr string
}

// Load reads and validates the server configuration.
func Load() (Config, error) {
	addr := strings.TrimSpace(os.Getenv("SHATRANJ_HTTP_ADDR"))
	if addr == "" {
		addr = defaultHTTPAddr
	}

	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return Config{}, fmt.Errorf("invalid SHATRANJ_HTTP_ADDR %q: %w", addr, err)
	}
	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return Config{}, fmt.Errorf("invalid port in SHATRANJ_HTTP_ADDR %q: %w", addr, err)
	}
	if portNumber == 0 && port != "0" {
		return Config{}, fmt.Errorf("invalid port in SHATRANJ_HTTP_ADDR %q", addr)
	}

	return Config{HTTPAddr: addr}, nil
}
