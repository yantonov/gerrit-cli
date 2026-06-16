package command

import (
	"encoding/json"
	"fmt"
)

var resolveChangeIDCommand = Command{
	usage:          "resolve-change-id <url>",
	requiresClient: true,
	run: func(client *GerritClient, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("ERROR: resolve-change-id requires <url>")
		}
		return client.resolveChangeID(args[0])
	},
}

func (c *GerritClient) resolveChangeID(urlStr string) error {
	changeNumber, err := extractChangeNumberFromURL(urlStr)
	if err != nil {
		return err
	}

	response, err := c.get(fmt.Sprintf("changes/%s", changeNumber))
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return err
	}

	if changeID, ok := data["change_id"].(string); ok {
		fmt.Println(changeID)
	}
	return nil
}
