package inject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fdrolshagen/jetter/internal"
	"github.com/fdrolshagen/jetter/internal/parser"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func Auth(requests *[]internal.Request, env parser.Environment) error {
	tokens := make(map[string]string)
	for _, request := range *requests {
		for key, value := range request.Headers {
			if key == "Authorization" {
				if strings.Contains(value, "{{$auth.token") {
					re := regexp.MustCompile(`\{\{\$auth\.token\("([^"]+)"\)}}`)
					matches := re.FindStringSubmatch(value)
					variable := matches[0]
					authId := matches[1]

					if len(matches) == 2 {
						auth, ok := env.Security.Auth[authId]
						if !ok {
							return fmt.Errorf("failed to find auth for authId=%s", authId)
						}

						token, ok := tokens[authId]
						if !ok {
							var err error
							token, err = GetToken(auth)
							if err != nil {
								return fmt.Errorf("failed to get token for authId=%s: %v\n", authId, err)
							}
							tokens[authId] = token
						}
						request.Headers[key] = strings.ReplaceAll(value, variable, token)
					}
				}
			}
		}
	}
	return nil
}

func GetToken(auth parser.AuthConfig) (string, error) {
	if auth.Type != "OAuth2" {
		return "", fmt.Errorf("unsupported auth type: %s", auth.Type)
	}

	form := url.Values{}
	form.Set("grant_type", strings.ToLower(auth.GrantType))
	form.Set("client_id", auth.ClientID)
	form.Set("client_secret", auth.ClientSecret)
	form.Set("username", auth.Username)
	form.Set("password", auth.Password)

	req, err := http.NewRequest("POST", auth.TokenURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed: %d %s", resp.StatusCode, string(body))
	}

	var tr TokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	return tr.AccessToken, nil
}
