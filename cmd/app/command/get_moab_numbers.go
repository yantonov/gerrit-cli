package command

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var getMoabNumbersCommand = Command{
	usage:          "get-moab-numbers <change_id>",
	requiresClient: true,
	run: func(client *GerritClient, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("ERROR: get-moab-numbers requires <change_id>")
		}
		return client.getMoabNumbers(args[0])
	},
}

func (c *GerritClient) getMoabNumbers(changeID string) error {
	response, err := c.get(fmt.Sprintf("changes/%s/messages", changeID))
	if err != nil {
		return err
	}

	var data []map[string]interface{}
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return err
	}

	moabNumbers := map[string]int{}
	pattern := regexp.MustCompile(`This review has been processed in (\w+) MOAB #(\d+)`)

	for _, comment := range data {
		if message, ok := comment["message"].(string); ok {
			matches := pattern.FindAllStringSubmatch(message, -1)
			for _, match := range matches {
				if len(match) >= 3 {
					moabType := match[1]
					var moabNumber int
					fmt.Sscanf(match[2], "%d", &moabNumber)
					moabNumbers[moabType] = moabNumber
				}
			}
		}
	}

	output, err := json.MarshalIndent(moabNumbers, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
