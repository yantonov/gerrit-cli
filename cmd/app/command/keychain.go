package command

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/term"
)

var keychainCommand = Command{
	usage:          "keychain <set|remove|clear|status|check> [username|password]",
	requiresClient: false,
	run: func(_ *GerritClient, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("ERROR: keychain requires a subcommand: set, remove, clear, status, check")
		}

		action, rest := args[0], args[1:]

		switch action {
		case "set":
			return keychainSet(rest)
		case "remove":
			return keychainRemove(rest)
		case "clear":
			return keychainClear(rest)
		case "status":
			return keychainStatus(rest)
		case "check":
			return keychainCheck(rest)
		default:
			return fmt.Errorf("ERROR: unknown keychain subcommand '%s'. Use: set, remove, clear, status, check", action)
		}
	},
}

func keychainField(name string) (string, error) {
	switch name {
	case keyringUsernameKey, keyringPasswordKey:
		return name, nil
	default:
		return "", fmt.Errorf("ERROR: field must be 'username' or 'password', got '%s'", name)
	}
}

func keychainSet(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("ERROR: keychain set requires exactly one field: username or password.\nValues are read interactively, not passed as arguments.\nRun: gerrit-cli keychain set <username|password>")
	}

	field, err := keychainField(args[0])
	if err != nil {
		return err
	}

	host, err := getGerritHost()
	if err != nil {
		return err
	}

	value, err := readHiddenValue(field)
	if err != nil {
		return err
	}

	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("ERROR: %s value is empty", field)
	}

	if err := setKeychainValue(host, field, value); err != nil {
		return keychainError(fmt.Sprintf("save %s to keychain", field), err)
	}

	fmt.Printf("[OK] %s saved to keychain for host %s\n", field, host)
	return nil
}

func keychainRemove(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("ERROR: keychain remove requires a field: username or password")
	}

	field, err := keychainField(args[0])
	if err != nil {
		return err
	}

	host, err := getGerritHost()
	if err != nil {
		return err
	}

	if err := deleteKeychainValue(host, field); err != nil {
		return keychainError(fmt.Sprintf("remove %s from keychain", field), err)
	}

	fmt.Printf("[OK] %s removed from keychain for host %s\n", field, host)
	return nil
}

func keychainClear(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("ERROR: keychain clear does not take arguments")
	}

	host, err := getGerritHost()
	if err != nil {
		return err
	}

	if err := deleteKeychainValue(host, keyringUsernameKey); err != nil {
		return keychainError("remove username from keychain", err)
	}
	if err := deleteKeychainValue(host, keyringPasswordKey); err != nil {
		return keychainError("remove password from keychain", err)
	}

	fmt.Printf("[OK] keychain cleared for host %s\n", host)
	return nil
}

func keychainStatus(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("ERROR: keychain status does not take arguments")
	}

	host, err := getGerritHost()
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

	fmt.Printf("Host: %s\n", host)
	fmt.Println("Username:", setStatus(username))
	fmt.Println("Password:", setStatus(password))

	return nil
}

func setStatus(value string) string {
	if value == "" {
		return "not set"
	}
	return "set"
}

// selfAccount is the part of the /accounts/self response used to name the
// account that the stored credentials belong to.
type selfAccount struct {
	Name     string `json:"name"`
	Username string `json:"username"`
}

func keychainCheck(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("ERROR: keychain check does not take arguments")
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

// readHiddenValue prompts for and reads a value from stdin. On a terminal,
// input is read without echoing, so the value never appears on screen.
func readHiddenValue(field string) (string, error) {
	fd := int(os.Stdin.Fd())

	if term.IsTerminal(fd) {
		fmt.Fprintf(os.Stderr, "%s%s: ", strings.ToUpper(field[:1]), field[1:])
		value, err := term.ReadPassword(fd)
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", fmt.Errorf("ERROR: failed to read %s: %v", field, err)
		}
		return string(value), nil
	}

	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("ERROR: failed to read %s: %v", field, err)
	}

	return strings.TrimRight(line, "\r\n"), nil
}
