package command

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

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
	if output != "[OK] Comment is successfully posted\n" {
		t.Fatalf("output = %q, want success message", output)
	}
}

func TestPostCommentReturnsStatusCodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "failed")
	}))
	defer server.Close()

	client := &GerritClient{auth: "token", baseURL: server.URL}
	output, err := captureStdout(t, func() error {
		return client.postComment("Iabc123", "Looks good")
	})
	if err == nil {
		t.Fatal("postComment returned nil error for non-200 response")
	}
	if err.Error() != "Cannot post comment, status code=500" {
		t.Fatalf("error = %q, want status code error", err)
	}
	if output != "" {
		t.Fatalf("output = %q, want empty stdout", output)
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
