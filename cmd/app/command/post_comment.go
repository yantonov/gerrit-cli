package command

import (
	"fmt"
	"net/http"
	"strings"
)

var reviewCommand = Command{
	usage:          "review (--change-id <change_id> | --review-url <url>) <message>",
	requiresClient: true,
	run: func(client *GerritClient, args []string) error {
		changeID, rest, err := parseChangeIDFlag("review", args)
		if err != nil {
			return err
		}
		if len(rest) < 1 {
			return fmt.Errorf("ERROR: review requires <message>")
		}
		return client.postReview(changeID, strings.Join(rest, " "))
	},
}

func (c *GerritClient) postReview(changeID, comment string) error {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return fmt.Errorf("ERROR: message is empty")
	}

	statusCode, err := c.post(
		fmt.Sprintf("changes/%s/revisions/current/review", changeID),
		map[string]string{"message": comment},
	)
	if err != nil {
		return err
	}

	if statusCode == http.StatusOK {
		fmt.Println("[OK] Review is successfully posted")
		return nil
	}

	return fmt.Errorf("Cannot post review, status code=%d", statusCode)
}
