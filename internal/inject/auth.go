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

	form, err := getFormValues(auth)
	if err != nil {
		return "", err
	}

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

func getFormValues(auth parser.AuthConfig) (url.Values, error) {
	form := url.Values{}

	switch auth.GrantType {
	case "Password":
		if auth.Username == "" || auth.Password == "" {
			return nil, fmt.Errorf("username and password required for password grant")
		}
		form.Set("grant_type", "password")
		form.Set("username", auth.Username)
		form.Set("password", auth.Password)
		form.Set("client_id", auth.ClientID)
		if auth.ClientSecret != "" {
			form.Set("client_secret", auth.ClientSecret)
		}

	case "Client Credentials":
		form.Set("grant_type", "client_credentials")
		form.Set("client_id", auth.ClientID)
		if auth.ClientSecret == "" {
			return nil, fmt.Errorf("client_secret required for client_credentials grant")
		}
		form.Set("client_secret", auth.ClientSecret)

	default:
		return nil, fmt.Errorf("unsupported grant type: %s", auth.GrantType)
	}

	if auth.Scope != "" {
		form.Set("scope", auth.Scope)
	}

	return form, nil
}
