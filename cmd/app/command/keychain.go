package command

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

var keychainCommand = Command{
	usage:          "keychain <set|remove|clear|status> [username|password]",
	requiresClient: false,
	run: func(_ *GerritClient, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("ERROR: keychain requires a subcommand: set, remove, clear, status")
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
		default:
			return fmt.Errorf("ERROR: unknown keychain subcommand '%s'. Use: set, remove, clear, status", action)
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
		return fmt.Errorf("ERROR: failed to save %s to keychain: %v", field, err)
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
		return fmt.Errorf("ERROR: failed to remove %s from keychain: %v", field, err)
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
		return fmt.Errorf("ERROR: failed to remove username from keychain: %v", err)
	}
	if err := deleteKeychainValue(host, keyringPasswordKey); err != nil {
		return fmt.Errorf("ERROR: failed to remove password from keychain: %v", err)
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
		return fmt.Errorf("ERROR: failed to read username from keychain: %v", err)
	}
	password, err := getKeychainValue(host, keyringPasswordKey)
	if err != nil {
		return fmt.Errorf("ERROR: failed to read password from keychain: %v", err)
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
