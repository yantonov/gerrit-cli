package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

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

func TestPostCommentPublishesReviewMessage(t *testing.T) {
	var gotMethod, gotPath, gotAuth, gotContentType, gotMessage string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read request body: %v", err)
			return
		}

		var payload map[string]string
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("failed to decode request body %q: %v", string(body), err)
			return
		}
		gotMessage = payload["message"]

		fmt.Fprint(w, ")]}'\n{\"ready\":true}")
	}))
	defer server.Close()

	client := &GerritClient{auth: "token", baseURL: server.URL}
	output, err := captureStdout(t, func() error {
		return client.postComment("Iabc123", "Looks good")
	})
	if err != nil {
		t.Fatalf("postComment returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q, want %q", gotMethod, http.MethodPost)
	}
	if gotPath != "/a/changes/Iabc123/revisions/current/review" {
		t.Fatalf("path = %q, want Gerrit review endpoint", gotPath)
	}
	if gotAuth != "Basic token" {
		t.Fatalf("Authorization = %q, want %q", gotAuth, "Basic token")
	}
	if gotContentType != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", gotContentType, "application/json")
	}
	if gotMessage != "Looks good" {
		t.Fatalf("message = %q, want %q", gotMessage, "Looks good")
	}
	if !strings.Contains(output, `"ready": true`) {
		t.Fatalf("output = %q, want formatted Gerrit response", output)
	}
}

func TestPostCommentRejectsEmptyComment(t *testing.T) {
	client := &GerritClient{auth: "token", baseURL: "http://example.com"}
	if err := client.postComment("Iabc123", " \t\n"); err == nil {
		t.Fatal("postComment returned nil error for empty comment")
	}
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	defer reader.Close()

	os.Stdout = writer
	callErr := fn()
	writer.Close()
	os.Stdout = oldStdout

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read stdout pipe: %v", err)
	}

	return string(output), callErr
}
