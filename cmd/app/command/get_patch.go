package command

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

var getPatchCommand = Command{
	usage:          "get-patch <change_id>",
	requiresClient: true,
	run: func(client *GerritClient, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("ERROR: get-patch requires <change_id>")
		}
		return client.getPatch(args[0])
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
