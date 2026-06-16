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

// TestExtractPublishVersionsAcrossMessages mirrors how getPublishVersion merges
// the per-message results of a full review thread. Only the "published" message
// contributes, so the presubmit MOAB_ID (1.1948302.1.1095021-review) and the
// publish-start version are ignored and the result is the published version.
func TestExtractPublishVersionsAcrossMessages(t *testing.T) {
	messages := []string{
		"Uploaded patch set 1.",
		"Patch Set 1: Code-Review-1",
		"CSHARP Presubmit [submitted](https://build-moab.crto.in/job/moabs/job/csharp-release/job/presubmit/job/MOAB-presubmit/1095021/)\nWaiting for available executor.",
		"Patch Set 1: Verified+1\n\n✅ **Presubmit succeeded: [All scans](https://build-scans.crto.in/scans?search.names=MOAB_ID&search.values=1.1948302.1.1095021-review)** - [Job](https://build-moab.crto.in/job/moabs/job/csharp-release/job/presubmit/job/MOAB-presubmit/1095021/)",
		"CSHARP Starting [publish of version 1.1948302.1.44265-review](https://build-moab.crto.in/job/moabs/job/csharp-release/job/presubmit/job/MOAB-presubmit-publish/44265/)\nWaiting for available executor.",
		"CSHARP Artifacts have been published in Nexus with version 1.1948302.1.44265-review",
	}

	merged := map[string]string{}
	for _, message := range messages {
		for lang, version := range extractPublishVersions(message) {
			merged[lang] = version
		}
	}

	want := map[string]string{"CSHARP": "1.1948302.1.44265-review"}
	if !reflect.DeepEqual(merged, want) {
		t.Fatalf("merged publish versions = %v, want %v", merged, want)
	}
}
