package command

import "testing"

func TestParseChangeIDFlagWithChangeID(t *testing.T) {
	changeID, rest, err := parseChangeIDFlag("get-change", []string{"--change-id", "Iabc123"})
	if err != nil {
		t.Fatalf("parseChangeIDFlag returned error: %v", err)
	}
	if changeID != "Iabc123" {
		t.Fatalf("changeID = %q, want %q", changeID, "Iabc123")
	}
	if len(rest) != 0 {
		t.Fatalf("rest = %v, want empty", rest)
	}
}

func TestParseChangeIDFlagWithChangeIDEqualsForm(t *testing.T) {
	changeID, _, err := parseChangeIDFlag("get-change", []string{"--change-id=Iabc123"})
	if err != nil {
		t.Fatalf("parseChangeIDFlag returned error: %v", err)
	}
	if changeID != "Iabc123" {
		t.Fatalf("changeID = %q, want %q", changeID, "Iabc123")
	}
}

func TestParseChangeIDFlagWithReviewURL(t *testing.T) {
	changeID, rest, err := parseChangeIDFlag(
		"get-diff",
		[]string{"--review-url", "https://gerrit.example.com/c/namespace/project/+/1234567", "path/to/file"},
	)
	if err != nil {
		t.Fatalf("parseChangeIDFlag returned error: %v", err)
	}
	if changeID != "1234567" {
		t.Fatalf("changeID = %q, want %q", changeID, "1234567")
	}
	if len(rest) != 1 || rest[0] != "path/to/file" {
		t.Fatalf("rest = %v, want [path/to/file]", rest)
	}
}

func TestParseChangeIDFlagWithReviewURLEqualsForm(t *testing.T) {
	changeID, _, err := parseChangeIDFlag(
		"get-change",
		[]string{"--review-url=https://gerrit.example.com/c/namespace/project/+/1234567"},
	)
	if err != nil {
		t.Fatalf("parseChangeIDFlag returned error: %v", err)
	}
	if changeID != "1234567" {
		t.Fatalf("changeID = %q, want %q", changeID, "1234567")
	}
}

func TestParseChangeIDFlagRequiresOneOfChangeIDOrReviewURL(t *testing.T) {
	if _, _, err := parseChangeIDFlag("get-change", []string{}); err == nil {
		t.Fatal("parseChangeIDFlag returned nil error when neither flag was given")
	}
}

func TestParseChangeIDFlagRejectsBothChangeIDAndReviewURL(t *testing.T) {
	_, _, err := parseChangeIDFlag(
		"get-change",
		[]string{"--change-id", "Iabc123", "--review-url", "https://gerrit.example.com/c/namespace/project/+/1234567"},
	)
	if err == nil {
		t.Fatal("parseChangeIDFlag returned nil error when both flags were given")
	}
}

func TestParseChangeIDFlagRejectsUnresolvableReviewURL(t *testing.T) {
	_, _, err := parseChangeIDFlag("get-change", []string{"--review-url", "https://gerrit.example.com/plugins/gitiles/project"})
	if err == nil {
		t.Fatal("parseChangeIDFlag returned nil error for a URL without a change number")
	}
}

func TestParseChangeIDFlagRejectsMissingValue(t *testing.T) {
	if _, _, err := parseChangeIDFlag("get-change", []string{"--change-id"}); err == nil {
		t.Fatal("parseChangeIDFlag returned nil error for --change-id with no value")
	}
	if _, _, err := parseChangeIDFlag("get-change", []string{"--review-url"}); err == nil {
		t.Fatal("parseChangeIDFlag returned nil error for --review-url with no value")
	}
}
