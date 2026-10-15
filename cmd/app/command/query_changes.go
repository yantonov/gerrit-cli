package command

import (
	"encoding/json"
	"fmt"
	"net/url"
)

var queryChangesCommand = Command{
	usage:          "query-changes --query <query>",
	requiresClient: true,
	run: func(client *GerritClient, args []string) error {
		query, _, err := parseQueryFlag("query-changes", args)
		if err != nil {
			return err
		}
		return client.queryChanges(query)
	},
}

func (c *GerritClient) queryChanges(query string) error {
	endpoint := fmt.Sprintf("changes/?q=%s", url.QueryEscape(query))
	response, err := c.get(endpoint)
	if err != nil {
		return err
	}

	var data []map[string]interface{}
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return err
	}

	results := make([]map[string]interface{}, 0, len(data))
	for _, change := range data {
		result := map[string]interface{}{
			"change_id": change["change_id"],
			"number":    change["_number"],
			"subject":   change["subject"],
			"project":   change["project"],
			"branch":    change["branch"],
			"status":    change["status"],
		}
		if owner, ok := change["owner"].(map[string]interface{}); ok {
			result["owner"] = owner["name"]
		}
		results = append(results, result)
	}

	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
