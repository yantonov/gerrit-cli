package command

import (
	"reflect"
	"testing"
)

func TestExtractPublishVersions(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    map[string]string
	}{
		{
			name:    "plain comment without a version",
			message: "test comment",
			want:    map[string]string{},
		},
		{
			name:    "presubmit submitted has a job number but no version",
			message: "CSHARP Presubmit [submitted](https://build-moab.crto.in/job/moabs/job/csharp-release/job/presubmit/job/MOAB-presubmit/1095021/)\nWaiting for available executor.",
			want:    map[string]string{},
		},
		{
			name:    "ownership message",
			message: "[Ownership] Owners for retail-media/catalog-api : `gu-retailmedia-catalog`",
			want:    map[string]string{},
		},
		{
			name:    "verified message is ignored (not a published message)",
			message: "Patch Set 1: Verified+1\n\n✅ **Presubmit succeeded: [All scans](https://build-scans.crto.in/scans?search.names=MOAB_ID&search.values=1.1948302.1.1095021-review)** - 📖 [Job](https://build-moab.crto.in/job/moabs/job/csharp-release/job/presubmit/job/MOAB-presubmit/1095021/)",
			want:    map[string]string{},
		},
		{
			name:    "publish start is ignored (says publish, not published)",
			message: "CSHARP Starting [publish of version 1.1948302.1.44265-review](https://build-moab.crto.in/job/moabs/job/csharp-release/job/presubmit/job/MOAB-presubmit-publish/44265/)\nWaiting for available executor.",
			want:    map[string]string{},
		},
		{
			name:    "artifacts published in nexus",
			message: "CSHARP Artifacts have been published in Nexus with version 1.1948302.1.44265-review\nhttps://s3.fr3.prod.crto.in/ci-summary-reports/moabs/csharp-release/presubmit/MOAB-presubmit-publish/44265/%2344265%20(retail-media/catalog-api%201.1948302.1.44265-review)/retail-media/catalog-api/publishedArtifacts/files/publishedArtifactLinks/publishedArtifactLinks.html",
			want:    map[string]string{"CSHARP": "1.1948302.1.44265-review"},
		},
		{
			name:    "published message without a language falls back to MOAB",
			message: "Artifacts have been published with version 1.1948302.1.44265-review",
			want:    map[string]string{"MOAB": "1.1948302.1.44265-review"},
		},
		{
			name:    "language casing is normalised to upper case",
			message: "JAVA Artifacts have been published in Nexus with version 2.5550123.1.999-review\nhttps://build-moab.crto.in/job/moabs/job/java-release/job/presubmit/",
			want:    map[string]string{"JAVA": "2.5550123.1.999-review"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := extractPublishVersions(test.message)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("extractPublishVersions() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestExtractVersionsAfterLastPleasePublish(t *testing.T) {
	tests := []struct {
		name     string
		messages []string
		want     map[string]string
	}{
		{
			name: "no please-publish returns empty",
			messages: []string{
				"Uploaded patch set 1.",
				"Patch Set 1: Code-Review-1",
				"CSHARP Presubmit [submitted](https://build-moab.crto.in/job/moabs/job/csharp-release/job/presubmit/job/MOAB-presubmit/1095021/)\nWaiting for available executor.",
				"Patch Set 1: Verified+1\n\n✅ **Presubmit succeeded: [All scans](https://build-scans.crto.in/scans?search.names=MOAB_ID&search.values=1.1948302.1.1095021-review)** - [Job](https://build-moab.crto.in/job/moabs/job/csharp-release/job/presubmit/job/MOAB-presubmit/1095021/)",
				"CSHARP Starting [publish of version 1.1948302.1.44265-review](https://build-moab.crto.in/job/moabs/job/csharp-release/job/presubmit/job/MOAB-presubmit-publish/44265/)\nWaiting for available executor.",
				"CSHARP Artifacts have been published in Nexus with version 1.1948302.1.44265-review",
			},
			want: map[string]string{},
		},
		{
			name: "versions before please-publish are ignored",
			messages: []string{
				"CSHARP Artifacts have been published in Nexus with version 1.1948302.1.11111-review",
				"please publish",
				"CSHARP Artifacts have been published in Nexus with version 1.1948302.1.44265-review",
			},
			want: map[string]string{"CSHARP": "1.1948302.1.44265-review"},
		},
		{
			name: "last please-publish wins when there are multiple",
			messages: []string{
				"please publish",
				"CSHARP Artifacts have been published in Nexus with version 1.1948302.1.11111-review",
				"please publish",
				"CSHARP Artifacts have been published in Nexus with version 1.1948302.1.44265-review",
			},
			want: map[string]string{"CSHARP": "1.1948302.1.44265-review"},
		},
		{
			name: "please-publish with no subsequent published message returns empty",
			messages: []string{
				"CSHARP Artifacts have been published in Nexus with version 1.1948302.1.11111-review",
				"please publish",
			},
			want: map[string]string{},
		},
		{
			name: "full review thread with please-publish",
			messages: []string{
				"Uploaded patch set 1.",
				"Patch Set 1: Code-Review-1",
				"CSHARP Presubmit [submitted](https://build-moab.crto.in/job/moabs/job/csharp-release/job/presubmit/job/MOAB-presubmit/1095021/)\nWaiting for available executor.",
				"Patch Set 1: Verified+1\n\n✅ **Presubmit succeeded: [All scans](https://build-scans.crto.in/scans?search.names=MOAB_ID&search.values=1.1948302.1.1095021-review)** - [Job](https://build-moab.crto.in/job/moabs/job/csharp-release/job/presubmit/job/MOAB-presubmit/1095021/)",
				"please publish",
				"CSHARP Starting [publish of version 1.1948302.1.44265-review](https://build-moab.crto.in/job/moabs/job/csharp-release/job/presubmit/job/MOAB-presubmit-publish/44265/)\nWaiting for available executor.",
				"CSHARP Artifacts have been published in Nexus with version 1.1948302.1.44265-review",
			},
			want: map[string]string{"CSHARP": "1.1948302.1.44265-review"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := extractVersionsAfterLastPleasePublish(test.messages)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("extractVersionsAfterLastPleasePublish() = %v, want %v", got, test.want)
			}
		})
	}
}
