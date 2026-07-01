package command

import (
	"fmt"
	"net/url"

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
	return keyring.Set(keyringServiceForHost(host), key, value)
}

// getKeychainValue returns "" (with no error) when the value has not been set.
func getKeychainValue(host, key string) (string, error) {
	value, err := keyring.Get(keyringServiceForHost(host), key)
	if err == keyring.ErrNotFound {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return value, nil
}

func deleteKeychainValue(host, key string) error {
	err := keyring.Delete(keyringServiceForHost(host), key)
	if err != nil && err != keyring.ErrNotFound {
		return err
	}

	return nil
}
