package command

import (
	"encoding/json"
	"fmt"
)

var getChangeCommand = Command{
	usage:          "get-change <change_id>",
	requiresClient: true,
	run: func(client *GerritClient, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("ERROR: get-change requires <change_id>")
		}
		return client.getChangeDetail(args[0])
	},
}

func (c *GerritClient) getChangeDetail(changeID string) error {
	response, err := c.get(fmt.Sprintf("changes/%s/detail", changeID))
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return err
	}

	result := map[string]interface{}{
		"subject": data["subject"],
		"project": data["project"],
		"branch":  data["branch"],
		"status":  data["status"],
	}

	if owner, ok := data["owner"].(map[string]interface{}); ok {
		result["owner"] = owner["name"]
	}

	if reviewers, ok := data["reviewers"].(map[string]interface{}); ok {
		if reviewerList, ok := reviewers["REVIEWER"].([]interface{}); ok {
			names := []string{}
			for _, r := range reviewerList {
				if reviewer, ok := r.(map[string]interface{}); ok {
					if name, ok := reviewer["name"].(string); ok {
						names = append(names, name)
					}
				}
			}
			result["reviewers"] = names
		}
	}

	result["labels"] = data["labels"]

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
