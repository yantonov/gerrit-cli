package command

import (
	"fmt"
	"strings"
)

// parseQueryFlag extracts the --query value from args.
func parseQueryFlag(commandName string, args []string) (string, []string, error) {
	var query string
	rest := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--query":
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("ERROR: %s: --query requires a value", commandName)
			}
			query = args[i+1]
			i++
		case strings.HasPrefix(arg, "--query="):
			query = strings.TrimPrefix(arg, "--query=")
		default:
			rest = append(rest, arg)
		}
	}

	if query == "" {
		return "", nil, fmt.Errorf("ERROR: %s requires --query", commandName)
	}

	return query, rest, nil
}
