package command

import (
	"reflect"
	"testing"
)

func TestExtractMoabNumbers(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    map[string]string
	}{
		{
			name:    "plain comment without a MOAB number",
			message: "test comment",
			want:    map[string]string{},
		},
		{
			name:    "single processed MOAB number",
			message: "This review has been processed in CSHARP MOAB #123",
			want:    map[string]string{"CSHARP": "123"},
		},
		{
			name:    "MOAB number embedded in a longer message",
			message: "Patch Set 1: Verified+1\n\nThis review has been processed in JAVA MOAB #4567\nDone.",
			want:    map[string]string{"JAVA": "4567"},
		},
		{
			name:    "multiple MOAB numbers in one message",
			message: "This review has been processed in CSHARP MOAB #1\nThis review has been processed in JAVA MOAB #2",
			want:    map[string]string{"CSHARP": "1", "JAVA": "2"},
		},
		{
			name:    "later MOAB number for the same type wins",
			message: "This review has been processed in CSHARP MOAB #1\nThis review has been processed in CSHARP MOAB #9",
			want:    map[string]string{"CSHARP": "9"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := extractMoabNumbers(test.message)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("extractMoabNumbers() = %v, want %v", got, test.want)
			}
		})
	}
}

// TestExtractMoabNumbersAcrossMessages mirrors how getMoabNumbers merges the
// per-message results of a full review thread.
func TestExtractMoabNumbersAcrossMessages(t *testing.T) {
	messages := []string{
		"Uploaded patch set 1.",
		"Patch Set 1: Code-Review-1",
		"This review has been processed in CSHARP MOAB #123",
		"This review has been processed in JAVA MOAB #456",
	}

	merged := map[string]string{}
	for _, message := range messages {
		for moabType, moabNumber := range extractMoabNumbers(message) {
			merged[moabType] = moabNumber
		}
	}

	want := map[string]string{"CSHARP": "123", "JAVA": "456"}
	if !reflect.DeepEqual(merged, want) {
		t.Fatalf("merged MOAB numbers = %v, want %v", merged, want)
	}
}
