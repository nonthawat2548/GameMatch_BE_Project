package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadJWTSecret(t *testing.T) {
	const fileSecret = "0123456789abcdef0123456789abcdef"
	const envSecret = "abcdef0123456789abcdef0123456789"
	for _, tc := range []struct {
		name    string
		env     *string
		file    string
		missing bool
		want    string
	}{
		{name: "file with newline", file: fileSecret + "\r\n", want: fileSecret},
		{name: "environment wins", env: ptrSecret(envSecret), file: fileSecret, want: envSecret},
		{name: "environment without file", env: ptrSecret(envSecret), missing: true, want: envSecret},
		{name: "explicit empty environment rejected", env: ptrSecret(""), file: fileSecret},
		{name: "short environment rejected", env: ptrSecret("short"), file: fileSecret},
		{name: "short file rejected", file: "short"},
		{name: "missing file rejected", missing: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("JWT_SECRET", "")
			if tc.env == nil {
				os.Unsetenv("JWT_SECRET")
			} else {
				os.Setenv("JWT_SECRET", *tc.env)
			}
			path := filepath.Join(t.TempDir(), ".jwt_secret")
			if !tc.missing {
				if err := os.WriteFile(path, []byte(tc.file), 0600); err != nil {
					t.Fatal(err)
				}
			}
			got, err := LoadJWTSecret(path)
			if tc.want == "" {
				if err == nil || got != "" {
					t.Fatal("expected rejection without a secret")
				}
				if strings.Contains(err.Error(), fileSecret) || strings.Contains(err.Error(), envSecret) {
					t.Fatal("error exposed a secret")
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatal("secret source or value was incorrect")
			}
		})
	}
}

func ptrSecret(value string) *string { return &value }
