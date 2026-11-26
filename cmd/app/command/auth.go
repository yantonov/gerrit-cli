package command

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var authCommand = Command{
	usage:          "auth <login>",
	requiresClient: false,
	run: func(_ *GerritClient, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("ERROR: auth requires a subcommand: login")
		}

		action, rest := args[0], args[1:]

		switch action {
		case "login":
			return authLogin(rest)
		default:
			return fmt.Errorf("ERROR: unknown auth subcommand '%s'. Use: login", action)
		}
	},
}

// selfAccount is the part of the /accounts/self response used to name the
// account that the stored credentials belong to.
type selfAccount struct {
	Name     string `json:"name"`
	Username string `json:"username"`
}

func authLogin(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("ERROR: auth login does not take arguments")
	}

	baseURL, err := getGerritURL()
	if err != nil {
		return err
	}

	host, err := hostFromBaseURL(baseURL)
	if err != nil {
		return err
	}

	username, err := getKeychainValue(host, keyringUsernameKey)
	if err != nil {
		return keychainError("read username from keychain", err)
	}
	password, err := getKeychainValue(host, keyringPasswordKey)
	if err != nil {
		return keychainError("read password from keychain", err)
	}

	missing := []string{}
	if username == "" {
		missing = append(missing, keyringUsernameKey)
	}
	if password == "" {
		missing = append(missing, keyringPasswordKey)
	}
	if len(missing) > 0 {
		return missingCredentialsError(host, missing)
	}

	client := &GerritClient{auth: basicAuth(username, password), baseURL: baseURL}
	body, status, err := client.getWithStatus("accounts/self")
	if err != nil {
		return fmt.Errorf("ERROR: failed to reach Gerrit at %s: %v", baseURL, err)
	}

	switch status {
	case http.StatusOK:
		fmt.Printf("[OK] login succeeded for host %s as %s\n", host, accountLabel(body, username))
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("ERROR: %s rejected the stored credentials for host %s.\nGet a new HTTP password at: %s/settings/#HTTPCredentials\nThen run: gerrit-cli keychain set password", baseURL, host, baseURL)
	case http.StatusForbidden:
		return fmt.Errorf("ERROR: %s accepted the credentials for host %s but denied access to the account details (HTTP 403)", baseURL, host)
	default:
		return fmt.Errorf("ERROR: unexpected response from %s (HTTP %d)", baseURL, status)
	}
}

func missingCredentialsError(host string, missing []string) error {
	commands := make([]string, 0, len(missing))
	for _, field := range missing {
		commands = append(commands, "Run: gerrit-cli keychain set "+field)
	}

	return fmt.Errorf("ERROR: no %s set for host %s.\n%s", strings.Join(missing, " and "), host, strings.Join(commands, "\n"))
}

// accountLabel names the account behind the credentials, falling back to the
// stored username when the server returns only an account id.
func accountLabel(body, username string) string {
	var account selfAccount
	if err := json.Unmarshal([]byte(body), &account); err != nil {
		return username
	}

	for _, label := range []string{account.Name, account.Username} {
		if label != "" {
			return label
		}
	}

	return username
}
