package command

import (
	"encoding/json"
	"fmt"
)

var getCommitCommand = Command{
	usage:          "get-commit <change_id>",
	requiresClient: true,
	run: func(client *GerritClient, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("ERROR: get-commit requires <change_id>")
		}
		return client.getCommit(args[0])
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
