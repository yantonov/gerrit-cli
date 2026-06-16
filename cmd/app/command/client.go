package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
