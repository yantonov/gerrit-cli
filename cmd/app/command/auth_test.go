package command

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestAuthLoginReportsMissingCredentials(t *testing.T) {
	tests := []struct {
		name         string
		stored       map[string]string
		wantContains string
	}{
		{
			name:         "username and password missing",
			stored:       nil,
			wantContains: "no username and password set for host gerrit.check-missing.example.com",
		},
		{
			name:         "only password stored",
			stored:       map[string]string{keyringPasswordKey: "secret"},
			wantContains: "no username set for host gerrit.check-missing.example.com",
		},
		{
			name:         "only username stored",
			stored:       map[string]string{keyringUsernameKey: "alice"},
			wantContains: "no password set for host gerrit.check-missing.example.com",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			keyring.MockInit()
			host := "gerrit.check-missing.example.com"
			t.Setenv("GERRIT_URL", host)

			for key, value := range test.stored {
				if err := setKeychainValue(host, key, value); err != nil {
					t.Fatalf("setKeychainValue() returned error: %v", err)
				}
			}

			err := authLogin(nil)
			if err == nil {
				t.Fatal("authLogin() returned nil error with incomplete credentials")
			}
			if !strings.Contains(err.Error(), test.wantContains) {
				t.Fatalf("error = %q, want it to contain %q", err, test.wantContains)
			}
		})
	}
}

func TestAuthLoginReportsLoginStatus(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		body           string
		wantErr        bool
		wantOutput     []string
		wantErrContain string
	}{
		{
			name:       "credentials accepted",
			status:     http.StatusOK,
			body:       ")]}'\n{\"_account_id\":1000,\"name\":\"Alice A\",\"username\":\"alice\"}",
			wantOutput: []string{"login succeeded for host", "as Alice A"},
		},
		{
			name:   "account id only",
			status: http.StatusOK,
			body:   ")]}'\n{\"_account_id\":1000}",
			// Without a name the stored username is the only label to show.
			wantOutput: []string{"login succeeded for host", "as alice"},
		},
		{
			name:           "credentials rejected",
			status:         http.StatusUnauthorized,
			body:           "Unauthorized",
			wantErr:        true,
			wantErrContain: "rejected the stored credentials",
		},
		{
			name:           "account details forbidden",
			status:         http.StatusForbidden,
			body:           "Forbidden",
			wantErr:        true,
			wantErrContain: "denied access to the account details",
		},
		{
			name:           "unexpected status",
			status:         http.StatusInternalServerError,
			body:           "boom",
			wantErr:        true,
			wantErrContain: "HTTP 500",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			keyring.MockInit()

			var gotPath, gotAuth string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotAuth = r.Header.Get("Authorization")

				w.WriteHeader(test.status)
				fmt.Fprint(w, test.body)
			}))
			defer server.Close()

			t.Setenv("GERRIT_URL", server.URL)
			// Credentials are keyed by the host GERRIT_URL resolves to, which for
			// a test server is the loopback address and port.
			host, err := hostFromBaseURL(server.URL)
			if err != nil {
				t.Fatalf("hostFromBaseURL() returned error: %v", err)
			}
			if err := setKeychainValue(host, keyringUsernameKey, "alice"); err != nil {
				t.Fatalf("setKeychainValue() returned error: %v", err)
			}
			if err := setKeychainValue(host, keyringPasswordKey, "secret"); err != nil {
				t.Fatalf("setKeychainValue() returned error: %v", err)
			}

			output, err := captureStdout(t, func() error {
				return authLogin(nil)
			})

			if gotPath != "/a/accounts/self" {
				t.Fatalf("path = %q, want %q", gotPath, "/a/accounts/self")
			}
			if gotAuth != "Basic YWxpY2U6c2VjcmV0" {
				t.Fatalf("Authorization = %q, want Basic auth of the stored credentials", gotAuth)
			}

			if test.wantErr {
				if err == nil {
					t.Fatalf("authLogin() returned nil error, want one containing %q", test.wantErrContain)
				}
				if !strings.Contains(err.Error(), test.wantErrContain) {
					t.Fatalf("error = %q, want it to contain %q", err, test.wantErrContain)
				}
				if strings.Contains(output, "login succeeded") {
					t.Fatalf("output = %q, must not report a successful login", output)
				}
				return
			}

			if err != nil {
				t.Fatalf("authLogin() returned error: %v", err)
			}
			for _, want := range test.wantOutput {
				if !strings.Contains(output, want) {
					t.Fatalf("output = %q, want it to contain %q", output, want)
				}
			}
			if strings.Contains(output, "secret") {
				t.Fatalf("output = %q, must not reveal the password", output)
			}
		})
	}
}

func TestAuthLoginRejectsArguments(t *testing.T) {
	if err := authLogin([]string{"extra"}); err == nil {
		t.Fatal("authLogin() returned nil error with an unexpected argument")
	}
}

// These fail on dispatch, before touching GERRIT_URL or the keychain.
func TestAuthRejectsBadSubcommands(t *testing.T) {
	cases := map[string][]string{
		"no subcommand":      {},
		"unknown subcommand": {"logout"},
	}

	for name, args := range cases {
		if err := authCommand.run(nil, args); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}
