package herdrsock

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	SocketPathEnv      = "HERDR_SOCKET_PATH"
	SessionEnv         = "HERDR_SESSION"
	DefaultSessionName = "default"
)

// ActiveSocketPath resolves the API socket using Herdr's environment contract:
// HERDR_SOCKET_PATH, HERDR_SESSION, then the default session socket.
func ActiveSocketPath() (string, error) {
	if path := os.Getenv(SocketPathEnv); path != "" {
		return path, nil
	}
	return SocketPathForSession(os.Getenv(SessionEnv))
}

// SocketPathForSession returns the default API socket path for a Herdr session.
func SocketPathForSession(name string) (string, error) {
	normalized, err := normalizeSessionName(name)
	if err != nil {
		return "", err
	}
	base := ConfigDir()
	if normalized == "" {
		return filepath.Join(base, "herdr.sock"), nil
	}
	return filepath.Join(base, "sessions", normalized, "herdr.sock"), nil
}

// ConfigDir returns Herdr's config directory for the current platform.
func ConfigDir() string {
	app := "herdr"
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, app)
	}
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, app)
		}
		if profile := os.Getenv("USERPROFILE"); profile != "" {
			return filepath.Join(profile, "AppData", "Roaming", app)
		}
	}
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config", app)
	}
	return filepath.Join(os.TempDir(), app)
}

func normalizeSessionName(name string) (string, error) {
	if name == "" || name == DefaultSessionName {
		return "", nil
	}
	if err := ValidateSessionName(name); err != nil {
		return "", err
	}
	return name, nil
}

// ValidateSessionName applies Herdr's named-session rules.
func ValidateSessionName(name string) error {
	if name == "" {
		return fmt.Errorf("herdrsock: session name cannot be empty")
	}
	if len(name) > 64 {
		return fmt.Errorf("herdrsock: session name cannot be longer than 64 bytes")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("herdrsock: session name cannot be . or ..")
	}
	for i := 0; i < len(name); i++ {
		b := name[i]
		if ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z') || ('0' <= b && b <= '9') || b == '.' || b == '_' || b == '-' {
			continue
		}
		return fmt.Errorf("herdrsock: session name may only contain ASCII letters, numbers, '.', '_' and '-'")
	}
	return nil
}
