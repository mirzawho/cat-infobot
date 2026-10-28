package telegram

import (
	"os"
	"path/filepath"
	"testing"
)

func writeEnv(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadConfigFromFile(t *testing.T) {
	path := writeEnv(t, "# comment\nexport BOT_TOKEN = \"123:abc\"\nTELEGRAM_API_ENDPOINT='https://example.test/'\nDATA_DIR=/tmp/x\n")
	t.Setenv("ENV_FILE", path)
	t.Setenv("BOT_TOKEN", "")
	t.Setenv("TELEGRAM_API_ENDPOINT", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Token != "123:abc" {
		t.Errorf("token = %q", cfg.Token)
	}
	if cfg.Endpoint != "https://example.test/" {
		t.Errorf("endpoint = %q", cfg.Endpoint)
	}
}

func TestEnvironmentWinsOverFile(t *testing.T) {
	path := writeEnv(t, "BOT_TOKEN=file-token\nTELEGRAM_API_ENDPOINT=https://file.test\n")
	t.Setenv("ENV_FILE", path)
	t.Setenv("BOT_TOKEN", "env-token")
	t.Setenv("TELEGRAM_API_ENDPOINT", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Token != "env-token" {
		t.Errorf("token = %q, want env value", cfg.Token)
	}
	if cfg.Endpoint != "https://file.test" {
		t.Errorf("endpoint = %q, want file value", cfg.Endpoint)
	}
}

func TestMissingToken(t *testing.T) {
	path := writeEnv(t, "TELEGRAM_API_ENDPOINT=\n")
	t.Setenv("ENV_FILE", path)
	t.Setenv("BOT_TOKEN", "")
	if _, err := LoadConfig(); err == nil {
		t.Fatal("expected an error when BOT_TOKEN is absent")
	}
	t.Setenv("ENV_FILE", filepath.Join(t.TempDir(), "missing.env"))
	if _, err := LoadConfig(); err == nil {
		t.Fatal("expected an error when the file is missing")
	}
}

func TestReadEnvFileIgnoresJunk(t *testing.T) {
	path := writeEnv(t, "not-a-pair\n\n# c\nA=1\nB = two words \nC=\n")
	vars, err := readEnvFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if vars["A"] != "1" || vars["B"] != "two words" || vars["C"] != "" {
		t.Errorf("vars = %#v", vars)
	}
	if _, ok := vars["not-a-pair"]; ok {
		t.Error("junk line parsed")
	}
}
