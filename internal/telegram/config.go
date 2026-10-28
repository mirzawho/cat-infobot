package telegram

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config is everything the bot needs to connect: a BotFather token and an
// optional Bot API endpoint.
type Config struct {
	Token    string
	Endpoint string
}

// LoadConfig reads BOT_TOKEN and TELEGRAM_API_ENDPOINT.
//
// Values come from the process environment first. Missing values are filled
// from .env in the working directory, so running ./bin/infobot directly works
// too. EnvFile overrides the .env path when set.
func LoadConfig() (Config, error) {
	file := os.Getenv("ENV_FILE")
	if file == "" {
		file = ".env"
	}
	vars, fileErr := readEnvFile(file)

	get := func(key string) string {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
		return strings.TrimSpace(vars[key])
	}

	cfg := Config{
		Token:    get("BOT_TOKEN"),
		Endpoint: get("TELEGRAM_API_ENDPOINT"),
	}
	if cfg.Token == "" {
		if fileErr != nil {
			return Config{}, fmt.Errorf("BOT_TOKEN is not set (and %s: %v)", file, fileErr)
		}
		return Config{}, fmt.Errorf("BOT_TOKEN is not set in the environment or %s", file)
	}
	return cfg, nil
}

// readEnvFile parses KEY=VALUE lines. Blank lines and # comments are skipped;
// an optional leading "export " and one pair of surrounding quotes are
// removed.
func readEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return map[string]string{}, err
	}
	defer f.Close()

	vars := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(value) >= 2 {
			if q := value[0]; (q == '"' || q == '\'') && value[len(value)-1] == q {
				value = value[1 : len(value)-1]
			}
		}
		if key != "" {
			vars[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return vars, err
	}
	return vars, nil
}

// DataDir is where the single-instance lock lives; a .env value wins over
// the default.
func DataDir() string {
	if v := strings.TrimSpace(os.Getenv("DATA_DIR")); v != "" {
		return v
	}
	if vars, err := readEnvFile(".env"); err == nil {
		if v := strings.TrimSpace(vars["DATA_DIR"]); v != "" {
			return v
		}
	}
	return "/data"
}
