package things

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	keychainService = "things-cli"
	keychainAccount = "auth-token"
	envVarName      = "THINGS_AUTH_TOKEN"
)

// ResolveAuthToken returns the auth token using the priority order:
// 1. Explicit token (from --auth-token flag)
// 2. THINGS_AUTH_TOKEN environment variable
// 3. macOS Keychain
func ResolveAuthToken(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	if token := os.Getenv(envVarName); token != "" {
		return token, nil
	}
	token, err := KeychainGet()
	if err != nil {
		return "", fmt.Errorf("no auth token found. Set THINGS_AUTH_TOKEN or run: things auth set")
	}
	return token, nil
}

// KeychainGet retrieves the auth token from macOS Keychain.
func KeychainGet() (string, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-s", keychainService,
		"-a", keychainAccount,
		"-w",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("keychain: %s", strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// KeychainSet stores the auth token in macOS Keychain.
func KeychainSet(token string) error {
	// Delete first to avoid "already exists" error
	_ = KeychainDelete()
	cmd := exec.Command("security", "add-generic-password",
		"-s", keychainService,
		"-a", keychainAccount,
		"-w", token,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("keychain set: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// KeychainDelete removes the auth token from macOS Keychain.
func KeychainDelete() error {
	cmd := exec.Command("security", "delete-generic-password",
		"-s", keychainService,
		"-a", keychainAccount,
	)
	return cmd.Run()
}

// AuthStatus returns a human-readable string describing the auth token status.
func AuthStatus() string {
	if token := os.Getenv(envVarName); token != "" {
		return fmt.Sprintf("Auth token: set via %s environment variable", envVarName)
	}
	if _, err := KeychainGet(); err == nil {
		return "Auth token: stored in macOS Keychain"
	}
	return "Auth token: not configured\n\nTo set up:\n  things auth set\n\nOr set the THINGS_AUTH_TOKEN environment variable.\n\nFind your token in Things > Settings > General > Enable Things URLs > Manage"
}
