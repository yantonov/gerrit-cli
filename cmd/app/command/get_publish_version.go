package command

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var getPublishVersionCommand = Command{
	usage:          "get-publish-version <change_id>",
	requiresClient: true,
	run: func(client *GerritClient, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("ERROR: get-publish-version requires <change_id>")
		}
		return client.getPublishVersion(args[0])
	},
}

func (c *GerritClient) getPublishVersion(changeID string) error {
	response, err := c.get(fmt.Sprintf("changes/%s/messages", changeID))
	if err != nil {
		return err
	}

	var data []map[string]interface{}
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return err
	}

	versions := map[string]string{}

	for _, comment := range data {
		if message, ok := comment["message"].(string); ok {
			for lang, version := range extractPublishVersions(message) {
				versions[lang] = version
			}
		}
	}

	output, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

// The CI bot no longer posts a free-text "CSHARP MOAB #123" phrase. Instead the
// published version now only shows up embedded in version / MOAB_ID strings, for
// example the message announcing that the artifacts have been published:
//
//	CSHARP Artifacts have been published in Nexus with version 1.1948302.1.44265-review
//	https://build-moab.crto.in/job/moabs/job/csharp-release/job/presubmit/...
//
// We only look at the "published" message: presubmit and publish-start messages
// quote other, non-final versions (e.g. the presubmit MOAB_ID
// 1.1948302.1.1095021-review) that we must not report.
//
// The published version is the whole "-review" version string (e.g.
// "1.1948302.1.44265-review"). Rather than matching the surrounding prose
// (which changes often), we anchor on the stable "-review" suffix and the
// dot-separated digits in front of it, which is far less fragile than the wording.

// publishedMessagePattern restricts extraction to the message that announces the
// published artifacts.
var publishedMessagePattern = regexp.MustCompile(`(?i)\bpublished\b`)

// publishVersionPattern captures the full published version, i.e. a run of
// dot-separated digits immediately followed by the "-review" suffix, such as
// "1.1948302.1.44265-review".
var publishVersionPattern = regexp.MustCompile(`\b\d[\d.]*-review\b`)

// The language/type is detected from two independent, redundant signals so a
// change to either one still leaves the other working:
//   - releaseJobPattern: the "<lang>-release" segment of a build-moab job
//     path, e.g. ".../job/csharp-release/..." -> "csharp".
//   - leadingTypePattern: the upper-case token the bot prefixes its
//     messages with, e.g. "CSHARP Artifacts have been published ..." -> "CSHARP".
var (
	releaseJobPattern  = regexp.MustCompile(`(?i)\b([a-z][a-z0-9]*)-release\b`)
	leadingTypePattern = regexp.MustCompile(`^([A-Z][A-Z0-9]+)\b`)
)

func extractPublishVersions(message string) map[string]string {
	versions := map[string]string{}

	if !publishedMessagePattern.MatchString(message) {
		return versions
	}

	version := publishVersionPattern.FindString(message)
	if version == "" {
		return versions
	}

	versions[extractPublishVersionType(message)] = version
	return versions
}

func extractPublishVersionType(message string) string {
	if match := releaseJobPattern.FindStringSubmatch(message); match != nil {
		return strings.ToUpper(match[1])
	}
	if match := leadingTypePattern.FindStringSubmatch(message); match != nil {
		return match[1]
	}
	return "MOAB"
}
