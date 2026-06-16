package command

import "testing"

func TestExtractChangeNumberFromURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "modern change URL",
			raw:  "https://gerrit.example.com/c/namespace/project/+/1234567",
			want: "1234567",
		},
		{
			name: "modern change URL with patch set",
			raw:  "https://gerrit.example.com/c/namespace/project/+/1234567/3",
			want: "1234567",
		},
		{
			name: "old hash change URL",
			raw:  "https://gerrit.example.com/#/c/namespace/project/+/1234567/3",
			want: "1234567",
		},
		{
			name: "old hash numeric change URL",
			raw:  "https://gerrit.example.com/#/c/1234567",
			want: "1234567",
		},
		{
			name: "encoded plus segment",
			raw:  "https://gerrit.example.com/c/namespace/project/%2B/1234567",
			want: "1234567",
		},
		{
			name: "numeric search URL",
			raw:  "https://gerrit.example.com/q/1234567",
			want: "1234567",
		},
		{
			name: "query URL",
			raw:  "https://gerrit.example.com/dashboard/?q=change%3A1234567",
			want: "1234567",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := extractChangeNumberFromURL(test.raw)
			if err != nil {
				t.Fatalf("extractChangeNumberFromURL returned error: %v", err)
			}
			if got != test.want {
				t.Fatalf("extractChangeNumberFromURL() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestExtractChangeNumberFromURLRejectsInvalidURL(t *testing.T) {
	if _, err := extractChangeNumberFromURL("https://gerrit.example.com/plugins/gitiles/project"); err == nil {
		t.Fatal("extractChangeNumberFromURL returned nil error for URL without change number")
	}
}

func TestExtractChangeNumberFromURLRejectsChangeIDURL(t *testing.T) {
	raw := "https://gerrit.example.com/q/I3ea8ccae945a1a1a0c52aab84bb1d2c1830bb2e3"
	if _, err := extractChangeNumberFromURL(raw); err == nil {
		t.Fatal("extractChangeNumberFromURL returned nil error for URL with commit Change-Id")
	}
}
