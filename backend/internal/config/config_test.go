package config_test

import (
	"testing"

	"github.com/hassaku63/chord-rambler/backend/internal/config"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		wantAddr string
		wantErr  bool
	}{
		{name: "PORT unset uses the default", port: "", wantAddr: ":8080"},
		{name: "PORT is used as given", port: "9000", wantAddr: ":9000"},
		{name: "lowest port", port: "1", wantAddr: ":1"},
		{name: "highest port", port: "65535", wantAddr: ":65535"},
		{name: "not a number", port: "http", wantErr: true},
		{name: "zero", port: "0", wantErr: true},
		{name: "above the range", port: "65536", wantErr: true},
		{name: "negative", port: "-1", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(key string) string {
				if key == "PORT" {
					return tt.port
				}
				return ""
			}

			got, err := config.Load(getenv)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Load() with PORT=%q = %+v, want an error", tt.port, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() with PORT=%q returned an error: %v", tt.port, err)
			}
			if got.Addr != tt.wantAddr {
				t.Errorf("Load() with PORT=%q: Addr = %q, want %q", tt.port, got.Addr, tt.wantAddr)
			}
		})
	}
}
