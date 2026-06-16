package command

import (
	"fmt"
	"net/http"
	"strings"
)

var postCommentCommand = Command{
	usage:          "post-comment <change_id> <comment>",
	requiresClient: true,
	run: func(client *GerritClient, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("ERROR: post-comment requires <change_id> <comment>")
		}
		return client.postComment(args[0], strings.Join(args[1:], " "))
	},
}

func (c *GerritClient) postComment(changeID, comment string) error {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return fmt.Errorf("ERROR: comment is empty")
	}

	statusCode, err := c.post(
		fmt.Sprintf("changes/%s/revisions/current/review", changeID),
		map[string]string{"message": comment},
	)
	if err != nil {
		return err
	}

	if statusCode == http.StatusOK {
		fmt.Println("[OK] Comment is successfully posted")
		return nil
	}

	return fmt.Errorf("Cannot post comment, status code=%d", statusCode)
}
