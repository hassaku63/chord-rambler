// Package config reads the server's settings from the environment.
package config

import (
	"fmt"
	"strconv"
)

// DefaultPort is the port the server listens on when PORT is not set.
const DefaultPort = 8080

// Config holds the server's settings.
type Config struct {
	// Addr is the address the HTTP server listens on, such as ":8080".
	Addr string
}

// Load builds a Config from getenv, which has the shape of os.Getenv.
// PORT is the only setting; it must be a number from 1 to 65535.
func Load(getenv func(string) string) (Config, error) {
	port := DefaultPort
	if raw := getenv("PORT"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 65535 {
			return Config{}, fmt.Errorf("config: PORT must be a number from 1 to 65535, got %q", raw)
		}
		port = n
	}
	return Config{Addr: ":" + strconv.Itoa(port)}, nil
}
