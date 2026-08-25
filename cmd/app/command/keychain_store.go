package command

import (
	"errors"
	"fmt"
	"net/url"
	"runtime"
	"strings"

	"github.com/zalando/go-keyring"
)

const (
	keyringUsernameKey = "username"
	keyringPasswordKey = "password"
)

func keyringServiceForHost(host string) string {
	return fmt.Sprintf("gerrit-cli:%s", host)
}

func hostFromBaseURL(baseURL string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("ERROR: failed to parse GERRIT_URL: %v", err)
	}

	return parsed.Host, nil
}

func getGerritHost() (string, error) {
	baseURL, err := getGerritURL()
	if err != nil {
		return "", err
	}

	return hostFromBaseURL(baseURL)
}

func setKeychainValue(host, key, value string) error {
	return asKeychainError(keyring.Set(keyringServiceForHost(host), key, value))
}

// getKeychainValue returns "" (with no error) when the value has not been set.
func getKeychainValue(host, key string) (string, error) {
	value, err := keyring.Get(keyringServiceForHost(host), key)
	if err == keyring.ErrNotFound {
		return "", nil
	}
	if err != nil {
		return "", asKeychainError(err)
	}

	return value, nil
}

func deleteKeychainValue(host, key string) error {
	err := keyring.Delete(keyringServiceForHost(host), key)
	if err != nil && err != keyring.ErrNotFound {
		return asKeychainError(err)
	}

	return nil
}

// keychainUnavailableError reports that no credential store could be reached at
// all, as opposed to a value simply being absent from a working one.
type keychainUnavailableError struct {
	cause error
}

func (e *keychainUnavailableError) Error() string {
	return fmt.Sprintf("ERROR: no keychain is available on this system (%s): %v\n%s", runtime.GOOS, e.cause, keychainSetupHint())
}

func (e *keychainUnavailableError) Unwrap() error {
	return e.cause
}

// unavailableKeychainMarkers are fragments the keyring backends report when the
// credential store itself is missing or unreachable.
var unavailableKeychainMarkers = []string{
	"org.freedesktop.secrets",   // no Secret Service provider is running
	"dbus",                      // no session bus to reach one through
	"executable file not found", // the macOS security binary is missing
	"no such file or directory",
}

// asKeychainError replaces a raw keyring failure with a keychainUnavailableError
// when the failure means the credential store cannot be reached at all.
func asKeychainError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, keyring.ErrUnsupportedPlatform) {
		return &keychainUnavailableError{cause: err}
	}

	message := strings.ToLower(err.Error())
	for _, marker := range unavailableKeychainMarkers {
		if strings.Contains(message, marker) {
			return &keychainUnavailableError{cause: err}
		}
	}

	return err
}

// keychainSetupHint explains, per platform, how to get a usable keychain.
func keychainSetupHint() string {
	switch runtime.GOOS {
	case "darwin":
		return "gerrit-cli stores credentials in the macOS Keychain via /usr/bin/security.\nMake sure that binary exists and the login keychain is unlocked, then retry."
	case "windows":
		return "gerrit-cli stores credentials in the Windows Credential Manager.\nMake sure the Credential Manager service (VaultSvc) is running, then retry."
	case "linux":
		return "gerrit-cli stores credentials in the freedesktop Secret Service.\nInstall and start a keyring daemon such as gnome-keyring or KeePassXC, and make sure a D-Bus session bus is running.\nOver SSH or in a container, try: dbus-run-session -- gerrit-cli <command>"
	default:
		return "gerrit-cli has no credential store backend for this platform."
	}
}

// keychainError wraps a keychain failure with the attempted action, keeping the
// actionable message intact when the keychain itself is unavailable.
func keychainError(action string, err error) error {
	var unavailable *keychainUnavailableError
	if errors.As(err, &unavailable) {
		return unavailable
	}

	return fmt.Errorf("ERROR: failed to %s: %v", action, err)
}
