package command

import (
	"fmt"
	"strings"
)

// parseQueryFlag extracts --query and --limit values from args.
func parseQueryFlag(commandName string, args []string) (string, string, error) {
	var query, limit string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--query":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("ERROR: %s: --query requires a value", commandName)
			}
			query = args[i+1]
			i++
		case strings.HasPrefix(arg, "--query="):
			query = strings.TrimPrefix(arg, "--query=")
		case arg == "--limit":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("ERROR: %s: --limit requires a value", commandName)
			}
			limit = args[i+1]
			i++
		case strings.HasPrefix(arg, "--limit="):
			limit = strings.TrimPrefix(arg, "--limit=")
		default:
			// unrecognized argument — skip
		}
	}

	if query == "" {
		return "", "", fmt.Errorf("ERROR: %s requires --query", commandName)
	}

	return query, limit, nil
}
