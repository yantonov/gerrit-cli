package command

import (
	"os"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestHostFromBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    string
	}{
		{name: "https URL", baseURL: "https://gerrit.example.com", want: "gerrit.example.com"},
		{name: "http URL", baseURL: "http://gerrit.example.com", want: "gerrit.example.com"},
		{name: "URL with port", baseURL: "https://gerrit.example.com:8080", want: "gerrit.example.com:8080"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := hostFromBaseURL(test.baseURL)
			if err != nil {
				t.Fatalf("hostFromBaseURL() returned error: %v", err)
			}
			if got != test.want {
				t.Fatalf("hostFromBaseURL() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestKeychainStoreRoundtrip(t *testing.T) {
	keyring.MockInit()

	host := "gerrit.roundtrip.example.com"

	if value, err := getKeychainValue(host, keyringUsernameKey); err != nil || value != "" {
		t.Fatalf("getKeychainValue() on unset key = (%q, %v), want (\"\", nil)", value, err)
	}

	if err := setKeychainValue(host, keyringUsernameKey, "alice"); err != nil {
		t.Fatalf("setKeychainValue() returned error: %v", err)
	}

	value, err := getKeychainValue(host, keyringUsernameKey)
	if err != nil {
		t.Fatalf("getKeychainValue() returned error: %v", err)
	}
	if value != "alice" {
		t.Fatalf("getKeychainValue() = %q, want %q", value, "alice")
	}

	if err := deleteKeychainValue(host, keyringUsernameKey); err != nil {
		t.Fatalf("deleteKeychainValue() returned error: %v", err)
	}

	if value, err := getKeychainValue(host, keyringUsernameKey); err != nil || value != "" {
		t.Fatalf("getKeychainValue() after delete = (%q, %v), want (\"\", nil)", value, err)
	}

	// Deleting an already-absent value must not error.
	if err := deleteKeychainValue(host, keyringUsernameKey); err != nil {
		t.Fatalf("deleteKeychainValue() on absent key returned error: %v", err)
	}
}

func TestKeychainSetUsernameReadsFromStdin(t *testing.T) {
	keyring.MockInit()
	t.Setenv("GERRIT_URL", "gerrit.set-username.example.com")

	var output string
	withStdinInput(t, "alice\n", func() {
		var err error
		output, err = captureStdout(t, func() error {
			return keychainSet([]string{"username"})
		})
		if err != nil {
			t.Fatalf("keychainSet() returned error: %v", err)
		}
	})
	if !strings.Contains(output, "username saved to keychain for host gerrit.set-username.example.com") {
		t.Fatalf("output = %q, want confirmation message", output)
	}

	value, err := getKeychainValue("gerrit.set-username.example.com", keyringUsernameKey)
	if err != nil {
		t.Fatalf("getKeychainValue() returned error: %v", err)
	}
	if value != "alice" {
		t.Fatalf("stored username = %q, want %q", value, "alice")
	}
}

func TestKeychainSetPasswordReadsFromStdin(t *testing.T) {
	keyring.MockInit()
	t.Setenv("GERRIT_URL", "gerrit.set-password.example.com")

	withStdinInput(t, "s3cret\n", func() {
		if err := keychainSet([]string{"password"}); err != nil {
			t.Fatalf("keychainSet() returned error: %v", err)
		}
	})

	value, err := getKeychainValue("gerrit.set-password.example.com", keyringPasswordKey)
	if err != nil {
		t.Fatalf("getKeychainValue() returned error: %v", err)
	}
	if value != "s3cret" {
		t.Fatalf("stored password = %q, want %q", value, "s3cret")
	}
}

func TestKeychainSetRejectsPositionalValue(t *testing.T) {
	keyring.MockInit()
	t.Setenv("GERRIT_URL", "gerrit.example.com")

	for _, field := range []string{"username", "password"} {
		err := keychainSet([]string{field, "value"})
		if err == nil {
			t.Fatalf("keychainSet() returned nil error for positional %s value", field)
		}
		if !strings.Contains(err.Error(), "read interactively") {
			t.Fatalf("error = %q, want mention of interactive entry", err)
		}
	}
}

func TestKeychainSetRejectsUnknownField(t *testing.T) {
	keyring.MockInit()
	t.Setenv("GERRIT_URL", "gerrit.example.com")

	withStdinInput(t, "value\n", func() {
		if err := keychainSet([]string{"token"}); err == nil {
			t.Fatal("keychainSet() returned nil error for unknown field")
		}
	})
}

func TestKeychainSetRejectsEmptyValue(t *testing.T) {
	keyring.MockInit()
	t.Setenv("GERRIT_URL", "gerrit.example.com")

	withStdinInput(t, "\n", func() {
		if err := keychainSet([]string{"username"}); err == nil {
			t.Fatal("keychainSet() returned nil error for empty username value")
		}
	})
}

func TestKeychainRemove(t *testing.T) {
	keyring.MockInit()
	t.Setenv("GERRIT_URL", "gerrit.remove.example.com")

	if err := setKeychainValue("gerrit.remove.example.com", keyringUsernameKey, "alice"); err != nil {
		t.Fatalf("setKeychainValue() returned error: %v", err)
	}

	if err := keychainRemove([]string{"username"}); err != nil {
		t.Fatalf("keychainRemove() returned error: %v", err)
	}

	value, err := getKeychainValue("gerrit.remove.example.com", keyringUsernameKey)
	if err != nil {
		t.Fatalf("getKeychainValue() returned error: %v", err)
	}
	if value != "" {
		t.Fatalf("username after remove = %q, want empty", value)
	}
}

func TestKeychainClearRemovesBothFields(t *testing.T) {
	keyring.MockInit()
	host := "gerrit.clear.example.com"
	t.Setenv("GERRIT_URL", host)

	if err := setKeychainValue(host, keyringUsernameKey, "alice"); err != nil {
		t.Fatalf("setKeychainValue() returned error: %v", err)
	}
	if err := setKeychainValue(host, keyringPasswordKey, "secret"); err != nil {
		t.Fatalf("setKeychainValue() returned error: %v", err)
	}

	if err := keychainClear(nil); err != nil {
		t.Fatalf("keychainClear() returned error: %v", err)
	}

	if value, _ := getKeychainValue(host, keyringUsernameKey); value != "" {
		t.Fatalf("username after clear = %q, want empty", value)
	}
	if value, _ := getKeychainValue(host, keyringPasswordKey); value != "" {
		t.Fatalf("password after clear = %q, want empty", value)
	}
}

func TestKeychainStatus(t *testing.T) {
	keyring.MockInit()
	host := "gerrit.status.example.com"
	t.Setenv("GERRIT_URL", host)

	output, err := captureStdout(t, func() error {
		return keychainStatus(nil)
	})
	if err != nil {
		t.Fatalf("keychainStatus() returned error: %v", err)
	}
	if !strings.Contains(output, "Username: not set") || !strings.Contains(output, "Password: not set") {
		t.Fatalf("output = %q, want both fields reported as not set", output)
	}

	if err := setKeychainValue(host, keyringUsernameKey, "alice"); err != nil {
		t.Fatalf("setKeychainValue() returned error: %v", err)
	}
	if err := setKeychainValue(host, keyringPasswordKey, "secret"); err != nil {
		t.Fatalf("setKeychainValue() returned error: %v", err)
	}

	output, err = captureStdout(t, func() error {
		return keychainStatus(nil)
	})
	if err != nil {
		t.Fatalf("keychainStatus() returned error: %v", err)
	}
	if !strings.Contains(output, "Username: set") {
		t.Fatalf("output = %q, want username reported as set", output)
	}
	if !strings.Contains(output, "Password: set") {
		t.Fatalf("output = %q, want password reported as set", output)
	}
	if strings.Contains(output, "alice") || strings.Contains(output, "secret") {
		t.Fatalf("output = %q, must not reveal the username or password value", output)
	}
}

func TestNewGerritClientErrorsWhenCredentialsMissing(t *testing.T) {
	keyring.MockInit()
	host := "gerrit.missing-creds.example.com"
	t.Setenv("GERRIT_URL", host)

	if _, err := NewGerritClient(); err == nil {
		t.Fatal("NewGerritClient() returned nil error with no credentials stored")
	} else if !strings.Contains(err.Error(), "keychain set username") {
		t.Fatalf("error = %q, want suggestion to run keychain set username", err)
	}

	if err := setKeychainValue(host, keyringUsernameKey, "alice"); err != nil {
		t.Fatalf("setKeychainValue() returned error: %v", err)
	}

	if _, err := NewGerritClient(); err == nil {
		t.Fatal("NewGerritClient() returned nil error with no password stored")
	} else if !strings.Contains(err.Error(), "keychain set password") {
		t.Fatalf("error = %q, want suggestion to run keychain set password", err)
	}
}

func TestNewGerritClientBuildsBasicAuthFromKeychain(t *testing.T) {
	keyring.MockInit()
	host := "gerrit.auth.example.com"
	t.Setenv("GERRIT_URL", host)

	if err := setKeychainValue(host, keyringUsernameKey, "alice"); err != nil {
		t.Fatalf("setKeychainValue() returned error: %v", err)
	}
	if err := setKeychainValue(host, keyringPasswordKey, "secret"); err != nil {
		t.Fatalf("setKeychainValue() returned error: %v", err)
	}

	client, err := NewGerritClient()
	if err != nil {
		t.Fatalf("NewGerritClient() returned error: %v", err)
	}

	const wantAuth = "YWxpY2U6c2VjcmV0" // base64("alice:secret")
	if client.auth != wantAuth {
		t.Fatalf("client.auth = %q, want %q", client.auth, wantAuth)
	}
}

// withStdinInput temporarily replaces os.Stdin with a pipe fed with the given
// input, so functions that read from stdin can be exercised in tests.
func withStdinInput(t *testing.T, input string, fn func()) {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}

	oldStdin := os.Stdin
	os.Stdin = reader
	defer func() {
		os.Stdin = oldStdin
		reader.Close()
	}()

	go func() {
		defer writer.Close()
		writer.WriteString(input)
	}()

	fn()
}
