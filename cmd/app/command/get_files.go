package command

import (
	"encoding/json"
	"fmt"
)

var getFilesCommand = Command{
	usage:          "get-files (--change-id <change_id> | --review-url <url>)",
	requiresClient: true,
	run: func(client *GerritClient, args []string) error {
		changeID, _, err := parseChangeIDFlag("get-files", args)
		if err != nil {
			return err
		}
		return client.getFiles(changeID)
	},
}

func (c *GerritClient) getFiles(changeID string) error {
	response, err := c.get(fmt.Sprintf("changes/%s/revisions/current/files", changeID))
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return err
	}

	keys := []string{}
	for k := range data {
		keys = append(keys, k)
	}

	output, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
