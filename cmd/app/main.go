package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type GerritClient struct {
	auth    string
	baseURL string
}

func normalizeBaseURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	rawURL = strings.TrimRight(rawURL, "/")

	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	return rawURL
}

func getGerritURL() (string, error) {
	baseURL := os.Getenv("GERRIT_URL")
	if baseURL == "" {
		return "", fmt.Errorf("ERROR: GERRIT_URL not set.\nSet your Gerrit base URL: export GERRIT_URL=your-gerrit-instance.com")
	}

	return normalizeBaseURL(baseURL), nil
}

func NewGerritClient() (*GerritClient, error) {
	auth := os.Getenv("SECRET_GERRIT_AUTH_TOKEN")
	if auth == "" {
		return nil, fmt.Errorf("ERROR: SECRET_GERRIT_AUTH_TOKEN not set.\nGet credentials from your Gerrit instance at: /settings/#HTTPCredentials\nThen run: export SECRET_GERRIT_AUTH_TOKEN=$(echo -n 'user:pass' | base64)")
	}

	baseURL, err := getGerritURL()
	if err != nil {
		return nil, err
	}

	return &GerritClient{auth: auth, baseURL: baseURL}, nil
}

func (c *GerritClient) get(endpoint string) (string, error) {
	url := fmt.Sprintf("%s/a/%s", c.baseURL, endpoint)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", c.auth))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	result := string(body)
	return stripGerritResponsePrefix(result), nil
}

func (c *GerritClient) post(endpoint string, payload interface{}) (int, error) {
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	url := fmt.Sprintf("%s/a/%s", c.baseURL, endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", c.auth))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if _, err := io.ReadAll(resp.Body); err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

func stripGerritResponsePrefix(result string) string {
	if strings.HasPrefix(result, ")]}'") {
		lines := strings.Split(result, "\n")
		if len(lines) > 1 {
			result = strings.Join(lines[1:], "\n")
		}
	}

	return result
}

func (c *GerritClient) postComment(changeID, comment string) error {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return fmt.Errorf("ERROR: comment is empty")
	}

	statusCode, err := c.post(
		fmt.Sprintf("changes/%s/revisions/current/review", changeID),
		map[string]string{"message": comment},
	)
	if err != nil {
		return err
	}

	if statusCode == http.StatusOK {
		fmt.Println("[OK] Comment is successfully posted")
		return nil
	}

	return fmt.Errorf("Cannot post comment, status code=%d", statusCode)
}

func (c *GerritClient) getChangeDetail(changeID string) error {
	response, err := c.get(fmt.Sprintf("changes/%s/detail", changeID))
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return err
	}

	result := map[string]interface{}{
		"subject": data["subject"],
		"project": data["project"],
		"branch":  data["branch"],
		"status":  data["status"],
	}

	if owner, ok := data["owner"].(map[string]interface{}); ok {
		result["owner"] = owner["name"]
	}

	if reviewers, ok := data["reviewers"].(map[string]interface{}); ok {
		if reviewerList, ok := reviewers["REVIEWER"].([]interface{}); ok {
			names := []string{}
			for _, r := range reviewerList {
				if reviewer, ok := r.(map[string]interface{}); ok {
					if name, ok := reviewer["name"].(string); ok {
						names = append(names, name)
					}
				}
			}
			result["reviewers"] = names
		}
	}

	result["labels"] = data["labels"]

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func (c *GerritClient) getFiles(changeID string) error {
	response, err := c.get(fmt.Sprintf("changes/%s/revisions/current/files", changeID))
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return err
	}

	keys := []string{}
	for k := range data {
		keys = append(keys, k)
	}

	output, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func (c *GerritClient) getCommit(changeID string) error {
	response, err := c.get(fmt.Sprintf("changes/%s/revisions/current/commit", changeID))
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return err
	}

	if message, ok := data["message"].(string); ok {
		fmt.Println(message)
	}
	return nil
}

func (c *GerritClient) getDiff(changeID, filePath string) error {
	encodedPath := url.PathEscape(filePath)
	response, err := c.get(fmt.Sprintf("changes/%s/revisions/current/files/%s/diff", changeID, encodedPath))
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return err
	}

	if content, ok := data["content"]; ok {
		output, err := json.MarshalIndent(content, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(output))
	}
	return nil
}

func (c *GerritClient) getMessages(changeID string) error {
	response, err := c.get(fmt.Sprintf("changes/%s/messages", changeID))
	if err != nil {
		return err
	}

	var data []map[string]interface{}
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return err
	}

	formattedComments := []map[string]interface{}{}
	for _, comment := range data {
		formatted := map[string]interface{}{
			"author":   "Unknown",
			"email":    "Unknown",
			"username": "Unknown",
			"message":  "",
		}

		if author, ok := comment["author"].(map[string]interface{}); ok {
			if name, ok := author["name"].(string); ok {
				formatted["author"] = name
			}
			if email, ok := author["email"].(string); ok {
				formatted["email"] = email
			}
			if username, ok := author["username"].(string); ok {
				formatted["username"] = username
			}
		}

		if message, ok := comment["message"].(string); ok {
			formatted["message"] = message
		}

		formattedComments = append(formattedComments, formatted)
	}

	output, err := json.MarshalIndent(formattedComments, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func (c *GerritClient) getPatch(changeID string) error {
	response, err := c.get(fmt.Sprintf("changes/%s/revisions/current/patch", changeID))
	if err != nil {
		return err
	}

	var encodedPatch string
	if err := json.Unmarshal([]byte(response), &encodedPatch); err != nil {
		return err
	}

	decoded, err := base64.StdEncoding.DecodeString(encodedPatch)
	if err != nil {
		return err
	}

	fmt.Println(string(decoded))
	return nil
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

	moabNumbers := map[string]int{}
	pattern := regexp.MustCompile(`This review has been processed in (\w+) MOAB #(\d+)`)

	for _, comment := range data {
		if message, ok := comment["message"].(string); ok {
			matches := pattern.FindAllStringSubmatch(message, -1)
			for _, match := range matches {
				if len(match) >= 3 {
					moabType := match[1]
					var moabNumber int
					fmt.Sscanf(match[2], "%d", &moabNumber)
					moabNumbers[moabType] = moabNumber
				}
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

func (c *GerritClient) resolveChangeID(urlStr string) error {
	changeNumber, err := extractChangeNumberFromURL(urlStr)
	if err != nil {
		return err
	}

	response, err := c.get(fmt.Sprintf("changes/%s", changeNumber))
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return err
	}

	if changeID, ok := data["change_id"].(string); ok {
		fmt.Println(changeID)
	}
	return nil
}

func extractChangeNumberFromURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", fmt.Errorf("ERROR: URL is empty")
	}

	candidates := []string{rawURL}
	if decodedURL, err := url.PathUnescape(rawURL); err == nil && decodedURL != rawURL {
		candidates = append(candidates, decodedURL)
	}

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?:^|[/?#])\+/(\d+)(?:[/?#&,]|$)`),
		regexp.MustCompile(`(?:^|[/?#])c/(\d+)(?:[/?#&,]|$)`),
		regexp.MustCompile(`(?:^|[/?#])q/(\d+)(?:[/?#&,]|$)`),
		regexp.MustCompile(`[?&]q=(?:change:)?(\d+)(?:[&#]|$)`),
	}

	for _, candidate := range candidates {
		for _, pattern := range patterns {
			match := pattern.FindStringSubmatch(candidate)
			if len(match) == 2 {
				return match[1], nil
			}
		}
	}

	return "", fmt.Errorf("ERROR: Could not extract change number from URL: %s", rawURL)
}

func resolveChangeNumber(urlStr string) error {
	changeNumber, err := extractChangeNumberFromURL(urlStr)
	if err != nil {
		return err
	}

	fmt.Println(changeNumber)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [args...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  get-change <change_id>\n")
		fmt.Fprintf(os.Stderr, "  get-files <change_id>\n")
		fmt.Fprintf(os.Stderr, "  get-commit <change_id>\n")
		fmt.Fprintf(os.Stderr, "  get-diff <change_id> <file_path>\n")
		fmt.Fprintf(os.Stderr, "  get-messages <change_id>\n")
		fmt.Fprintf(os.Stderr, "  get-patch <change_id>\n")
		fmt.Fprintf(os.Stderr, "  get-moab-numbers <change_id>\n")
		fmt.Fprintf(os.Stderr, "  post-comment <change_id> <comment>\n")
		fmt.Fprintf(os.Stderr, "  resolve-change-number <url>\n")
		fmt.Fprintf(os.Stderr, "  resolve-change-id <url>\n")
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	if command == "resolve-change-number" {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "ERROR: resolve-change-number requires <url>")
			os.Exit(1)
		}
		if err := resolveChangeNumber(args[0]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	client, err := NewGerritClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch command {
	case "get-change":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "ERROR: get-change requires <change_id>")
			os.Exit(1)
		}
		err = client.getChangeDetail(args[0])
	case "get-files":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "ERROR: get-files requires <change_id>")
			os.Exit(1)
		}
		err = client.getFiles(args[0])
	case "get-commit":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "ERROR: get-commit requires <change_id>")
			os.Exit(1)
		}
		err = client.getCommit(args[0])
	case "get-diff":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "ERROR: get-diff requires <change_id> <file_path>")
			os.Exit(1)
		}
		err = client.getDiff(args[0], args[1])
	case "get-messages":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "ERROR: get-messages requires <change_id>")
			os.Exit(1)
		}
		err = client.getMessages(args[0])
	case "get-patch":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "ERROR: get-patch requires <change_id>")
			os.Exit(1)
		}
		err = client.getPatch(args[0])
	case "get-moab-numbers":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "ERROR: get-moab-numbers requires <change_id>")
			os.Exit(1)
		}
		err = client.getMoabNumbers(args[0])
	case "post-comment":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "ERROR: post-comment requires <change_id> <comment>")
			os.Exit(1)
		}
		err = client.postComment(args[0], strings.Join(args[1:], " "))
	case "resolve-change-id":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "ERROR: resolve-change-id requires <url>")
			os.Exit(1)
		}
		err = client.resolveChangeID(args[0])
	default:
		fmt.Fprintf(os.Stderr, "ERROR: Unknown command '%s'\n", command)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
