package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	t.Setenv("GM_ENV_PRIORITY", "hosting-value")
	t.Setenv("GM_ENV_EMPTY", "")
	const localKey = "GM_ENV_LOCAL_TEST"
	old, existed := os.LookupEnv(localKey)
	os.Unsetenv(localKey)
	t.Cleanup(func() {
		if existed {
			os.Setenv(localKey, old)
		} else {
			os.Unsetenv(localKey)
		}
	})
	if err := os.WriteFile(path, []byte("# local configuration\nGM_ENV_PRIORITY=local-value\nGM_ENV_EMPTY=local-value\nGM_ENV_LOCAL_TEST='local-key'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := LoadEnv(path); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("GM_ENV_PRIORITY") != "hosting-value" || os.Getenv("GM_ENV_EMPTY") != "" {
		t.Fatal("local configuration overrode process environment")
	}
	if os.Getenv(localKey) != "local-key" {
		t.Fatal("local configuration was not loaded")
	}
}

func TestLoadEnvMissingAndInvalid(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := LoadEnv(path); err != nil {
		t.Fatal("missing optional file should not prevent startup")
	}
	if err := os.WriteFile(path, []byte("GM_ENV_BROKEN='unterminated\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := LoadEnv(path); err == nil {
		t.Fatal("invalid configuration should be rejected")
	}
}
