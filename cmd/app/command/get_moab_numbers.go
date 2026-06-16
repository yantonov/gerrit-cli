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

	moabNumbers := map[string]string{}

	for _, comment := range data {
		if message, ok := comment["message"].(string); ok {
			for moabType, moabNumber := range extractMoabNumbers(message) {
				moabNumbers[moabType] = moabNumber
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

// moabNumberPattern matches the CI bot's free-text announcement, e.g.
// "This review has been processed in CSHARP MOAB #123".
var moabNumberPattern = regexp.MustCompile(`This review has been processed in (\w+) MOAB #(\d+)`)

func extractMoabNumbers(message string) map[string]string {
	moabNumbers := map[string]string{}

	for _, match := range moabNumberPattern.FindAllStringSubmatch(message, -1) {
		if len(match) >= 3 {
			moabNumbers[match[1]] = match[2]
		}
	}

	return moabNumbers
}
