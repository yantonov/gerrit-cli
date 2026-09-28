package command

import (
	"encoding/json"
	"fmt"
)

var isVerifiedCommand = Command{
	usage:          "is-verified (--change-id <change_id> | --review-url <url>)",
	requiresClient: true,
	run: func(client *GerritClient, args []string) error {
		changeID, _, err := parseChangeIDFlag("is-verified", args)
		if err != nil {
			return err
		}
		return client.isVerified(changeID)
	},
}

func (c *GerritClient) isVerified(changeID string) error {
	response, err := c.get(fmt.Sprintf("changes/%s/detail", changeID))
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return err
	}

	labels, _ := data["labels"].(map[string]interface{})
	verifiedLabel, _ := labels["Verified"].(map[string]interface{})

	result := verifiedStatus(verifiedLabel)

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

// verifiedStatus interprets the Verified label block from a change's
// "detail" response. Gerrit resets non-sticky labels (Verified is one) on
// every new patch set, so this label already reflects only the current,
// latest revision - no need to correlate votes against a patch set number.
func verifiedStatus(verifiedLabel map[string]interface{}) map[string]interface{} {
	if verifiedLabel == nil {
		return map[string]interface{}{
			"status": "not-configured",
		}
	}

	if approver, ok := verifiedLabel["approved"].(map[string]interface{}); ok {
		return map[string]interface{}{
			"status":   "verified",
			"verified": true,
			"by":       approver["name"],
		}
	}

	if rejecter, ok := verifiedLabel["rejected"].(map[string]interface{}); ok {
		return map[string]interface{}{
			"status":   "rejected",
			"verified": false,
			"by":       rejecter["name"],
		}
	}

	return map[string]interface{}{
		"status":   "no-score",
		"verified": false,
	}
}
