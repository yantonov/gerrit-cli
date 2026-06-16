package command

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var resolveChangeNumberCommand = Command{
	usage:          "resolve-change-number <url>",
	requiresClient: false,
	run: func(_ *GerritClient, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("ERROR: resolve-change-number requires <url>")
		}
		return resolveChangeNumber(args[0])
	},
}

func extractChangeNumberFromURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", fmt.Errorf("ERROR: URL is empty")
	}

	candidates := []string{rawURL}
	if decodedURL, err := url.PathUnescape(rawURL); err == nil && decodedURL != rawURL {
		candidates = append(candidates, decodedURL)
	}

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?:^|[/?#])\+/(\d+)(?:[/?#&,]|$)`),
		regexp.MustCompile(`(?:^|[/?#])c/(\d+)(?:[/?#&,]|$)`),
		regexp.MustCompile(`(?:^|[/?#])q/(\d+)(?:[/?#&,]|$)`),
		regexp.MustCompile(`[?&]q=(?:change:)?(\d+)(?:[&#]|$)`),
	}

	for _, candidate := range candidates {
		for _, pattern := range patterns {
			match := pattern.FindStringSubmatch(candidate)
			if len(match) == 2 {
				return match[1], nil
			}
		}
	}

	return "", fmt.Errorf("ERROR: Could not extract change number from URL: %s", rawURL)
}

func resolveChangeNumber(urlStr string) error {
	changeNumber, err := extractChangeNumberFromURL(urlStr)
	if err != nil {
		return err
	}

	fmt.Println(changeNumber)
	return nil
}
