// Package config manages local CLI state: site, session and output prefs.
//
// Storage follows the XDG convention, same as gh:
//
//	$XDG_CONFIG_HOME/coj/config.json (default ~/.config/coj/config.json)
//	$XDG_CONFIG_HOME/coj/session      (session id, mode 0600)
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config is the persisted user configuration.
type Config struct {
	Host     string `json:"host"`      // site URL, e.g. https://candyoj.com
	Username string `json:"username"`  // last used account name
	UserType int    `json:"user_type"` // F_Type
	UserID   string `json:"user_id"`   // remembered F_UserID, used to infer context
	ClassID  string `json:"class_id"`  // remembered F_ClassID, used to infer context: 1 admin, 2 school-master, 3 teacher, 4 student
	Timeout  int    `json:"timeout"`   // timeout in seconds
}

// Dir returns the config directory.
func Dir() string {
	if d := os.Getenv("COJ_CONFIG_DIR"); d != "" {
		return d
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ".coj"
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "coj")
}

// Path returns the config file path.
func Path() string { return filepath.Join(Dir(), "config.json") }

// SessionPath returns the session file path.
func SessionPath() string { return filepath.Join(Dir(), "session") }

// Load reads the config, falling back to defaults.
func Load() (*Config, error) {
	cfg := &Config{Host: "https://candyoj.com", Timeout: 30}
	raw, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := json.Unmarshal(raw, cfg); err != nil {
		return cfg, fmt.Errorf("config file is corrupted: %w", err)
	}
	if cfg.Host == "" {
		cfg.Host = "https://candyoj.com"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30
	}
	return cfg, nil
}

// Save writes the config back.
func (c *Config) Save() error {
	if err := os.MkdirAll(Dir(), 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), raw, 0o600)
}

// SaveSession persists the session id.
func SaveSession(id string) error {
	if err := os.MkdirAll(Dir(), 0o700); err != nil {
		return err
	}
	return os.WriteFile(SessionPath(), []byte(strings.TrimSpace(id)), 0o600)
}

// LoadSession reads the session id.
func LoadSession() string {
	raw, err := os.ReadFile(SessionPath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(raw))
}

// ClearSession removes the session.
func ClearSession() error {
	if err := os.Remove(SessionPath()); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// RoleName turns F_Type into a role name.
func RoleName(t int) string {
	switch t {
	case 1:
		return "admin"
	case 2:
		return "school-master"
	case 3:
		return "teacher"
	case 4:
		return "student"
	}
	return fmt.Sprintf("unknown(%d)", t)
}
