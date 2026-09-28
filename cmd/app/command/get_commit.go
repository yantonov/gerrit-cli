package command

import (
	"encoding/json"
	"fmt"
)

var getCommitCommand = Command{
	usage:          "get-commit (--change-id <change_id> | --review-url <url>)",
	requiresClient: true,
	run: func(client *GerritClient, args []string) error {
		changeID, _, err := parseChangeIDFlag("get-commit", args)
		if err != nil {
			return err
		}
		return client.getCommit(changeID)
	},
}

func (c *GerritClient) getCommit(changeID string) error {
	response, err := c.get(fmt.Sprintf("changes/%s/revisions/current/commit", changeID))
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return err
	}

	if message, ok := data["message"].(string); ok {
		fmt.Println(message)
	}
	return nil
}
