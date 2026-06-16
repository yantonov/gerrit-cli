package command

import (
	"encoding/json"
	"fmt"
	"net/url"
)

var getDiffCommand = Command{
	usage:          "get-diff <change_id> <file_path>",
	requiresClient: true,
	run: func(client *GerritClient, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("ERROR: get-diff requires <change_id> <file_path>")
		}
		return client.getDiff(args[0], args[1])
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
