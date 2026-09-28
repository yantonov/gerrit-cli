package command

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

var getPatchCommand = Command{
	usage:          "get-patch (--change-id <change_id> | --review-url <url>)",
	requiresClient: true,
	run: func(client *GerritClient, args []string) error {
		changeID, _, err := parseChangeIDFlag("get-patch", args)
		if err != nil {
			return err
		}
		return client.getPatch(changeID)
	},
}

func (c *GerritClient) getPatch(changeID string) error {
	response, err := c.get(fmt.Sprintf("changes/%s/revisions/current/patch", changeID))
	if err != nil {
		return err
	}

	var encodedPatch string
	if err := json.Unmarshal([]byte(response), &encodedPatch); err != nil {
		return err
	}

	decoded, err := base64.StdEncoding.DecodeString(encodedPatch)
	if err != nil {
		return err
	}

	fmt.Println(string(decoded))
	return nil
}
