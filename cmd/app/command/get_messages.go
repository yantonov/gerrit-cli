package command

import (
	"encoding/json"
	"fmt"
)

var getMessagesCommand = Command{
	usage:          "get-messages <change_id>",
	requiresClient: true,
	run: func(client *GerritClient, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("ERROR: get-messages requires <change_id>")
		}
		return client.getMessages(args[0])
	},
}

func (c *GerritClient) getMessages(changeID string) error {
	response, err := c.get(fmt.Sprintf("changes/%s/messages", changeID))
	if err != nil {
		return err
	}

	var data []map[string]interface{}
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return err
	}

	formattedComments := []map[string]interface{}{}
	for _, comment := range data {
		formatted := map[string]interface{}{
			"author":   "Unknown",
			"email":    "Unknown",
			"username": "Unknown",
			"message":  "",
		}

		if author, ok := comment["author"].(map[string]interface{}); ok {
			if name, ok := author["name"].(string); ok {
				formatted["author"] = name
			}
			if email, ok := author["email"].(string); ok {
				formatted["email"] = email
			}
			if username, ok := author["username"].(string); ok {
				formatted["username"] = username
			}
		}

		if message, ok := comment["message"].(string); ok {
			formatted["message"] = message
		}

		formattedComments = append(formattedComments, formatted)
	}

	output, err := json.MarshalIndent(formattedComments, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
