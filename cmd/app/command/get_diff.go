package command

import (
	"encoding/json"
	"fmt"
	"net/url"
)

var getDiffCommand = Command{
	usage:          "get-diff (--change-id <change_id> | --review-url <url>) <file_path>",
	requiresClient: true,
	run: func(client *GerritClient, args []string) error {
		changeID, rest, err := parseChangeIDFlag("get-diff", args)
		if err != nil {
			return err
		}
		if len(rest) < 1 {
			return fmt.Errorf("ERROR: get-diff requires <file_path>")
		}
		return client.getDiff(changeID, rest[0])
	},
}

func (c *GerritClient) getDiff(changeID, filePath string) error {
	encodedPath := url.PathEscape(filePath)
	response, err := c.get(fmt.Sprintf("changes/%s/revisions/current/files/%s/diff", changeID, encodedPath))
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return err
	}

	if content, ok := data["content"]; ok {
		output, err := json.MarshalIndent(content, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(output))
	}
	return nil
}
