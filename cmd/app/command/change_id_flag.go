package command

import (
	"fmt"
	"strings"
)

// parseChangeIDFlag extracts a change identifier from --change-id or --review-url
// flags (either "--flag value" or "--flag=value" form), returning the resolved
// change ID and the remaining, non-flag args in their original order.
//
// Exactly one of --change-id or --review-url must be supplied. When --review-url
// is given, the change number is resolved from the URL and used as the change ID.
func parseChangeIDFlag(commandName string, args []string) (string, []string, error) {
	var changeID, reviewURL string
	rest := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--change-id":
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("ERROR: %s: --change-id requires a value", commandName)
			}
			changeID = args[i+1]
			i++
		case strings.HasPrefix(arg, "--change-id="):
			changeID = strings.TrimPrefix(arg, "--change-id=")
		case arg == "--review-url":
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("ERROR: %s: --review-url requires a value", commandName)
			}
			reviewURL = args[i+1]
			i++
		case strings.HasPrefix(arg, "--review-url="):
			reviewURL = strings.TrimPrefix(arg, "--review-url=")
		default:
			rest = append(rest, arg)
		}
	}

	if changeID == "" && reviewURL == "" {
		return "", nil, fmt.Errorf("ERROR: %s requires --change-id or --review-url", commandName)
	}
	if changeID != "" && reviewURL != "" {
		return "", nil, fmt.Errorf("ERROR: %s: specify either --change-id or --review-url, not both", commandName)
	}

	if reviewURL != "" {
		resolved, err := extractChangeNumberFromURL(reviewURL)
		if err != nil {
			return "", nil, fmt.Errorf("ERROR: %s: could not resolve change id from --review-url: %v", commandName, err)
		}
		changeID = resolved
	}

	return changeID, rest, nil
}
