package parser

import (
	"encoding/json"
	"github.com/fdrolshagen/jetter/internal"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestParseEnv(t *testing.T) {
	t.Run("invalid format", func(t *testing.T) {
		_, err := ParseEnv("invalidformat")
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid environment format")
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := ParseEnv("nofile.json:dev")
		assert.Error(t, err)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		tmp := filepath.Join(t.TempDir(), "invalid.json")
		err := os.WriteFile(tmp, []byte("{not json"), 0644)
		assert.NoError(t, err)

		_, err = ParseEnv(tmp + ":dev")
		assert.Error(t, err)
		assert.EqualError(t, err, "invalid environment content")
	})

	t.Run("valid config", func(t *testing.T) {
		jsonData := []byte(`{
		"dev": {
			"URL": "http://localhost:8080",
			"Security": {
				"Auth": {
					"auth-id": {
						"Type": "OAuth2",
						"Token URL": "http://localhost:8081/realms/test-realm/protocol/openid-connect/token",
						"Grant Type": "Password",
						"Client ID": "test-client",
						"Client Secret": "test-secret",
						"Username": "test-user",
						"Password": "test-password",
						"Scope": "read write"
					}
				}
			}
		}
	}`)

		tmp := filepath.Join(t.TempDir(), "valid.json")
		err := os.WriteFile(tmp, jsonData, 0644)
		assert.NoError(t, err)

		result, err := ParseEnv(tmp + ":dev")
		assert.NoError(t, err)
		assert.Equal(t, "http://localhost:8080", result.Variables["URL"])

		auth, ok := result.Security.Auth["auth-id"]
		assert.True(t, ok)
		assert.Equal(t, "OAuth2", auth.Type)
		assert.Equal(t, "Password", auth.GrantType)
		assert.Equal(t, "test-user", auth.Username)
	})

	t.Run("environment not found", func(t *testing.T) {
		cfg := internal.Config{"prod": internal.Environment{}}
		data, _ := json.Marshal(cfg)
		tmp := filepath.Join(t.TempDir(), "envnotfound.json")
		err := os.WriteFile(tmp, data, 0644)
		assert.NoError(t, err)

		result, err := ParseEnv(tmp + ":dev")
		assert.Error(t, err)
		assert.ErrorContains(t, err, "environment not found")
		assert.Empty(t, result.Variables)
	})
}
