package inject

import (
	"fmt"
	"github.com/fdrolshagen/jetter/internal"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestAuth_ReturnsErrorForMalformedPlaceholder(t *testing.T) {
	requests := []internal.Request{
		{
			Headers: map[string]string{
				"Authorization": "Bearer {{$auth.token(auth-id)}}",
			},
		},
	}

	err := Auth(&requests, internal.Environment{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid auth token placeholder")
}

func TestAuth_ReturnsErrorWhenAuthConfigMissing(t *testing.T) {
	requests := []internal.Request{
		{
			Headers: map[string]string{
				"Authorization": "Bearer {{$auth.token(\"missing-id\")}}",
			},
		},
	}

	env := internal.Environment{
		Security: internal.Security{Auth: internal.AuthMap{}},
	}

	err := Auth(&requests, env)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find auth for authId=missing-id")
}

func TestAuth_InjectsTokenAndCachesPerAuthID(t *testing.T) {
	var tokenRequests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&tokenRequests, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"access_token":"abc-token","refresh_token":"","expires_in":300,"token_type":"Bearer"}`)
	}))
	defer server.Close()

	requests := []internal.Request{
		{
			Headers: map[string]string{
				"Authorization": "Bearer {{$auth.token(\"auth-id\")}}",
			},
		},
		{
			Headers: map[string]string{
				"Authorization": "Bearer {{$auth.token(\"auth-id\")}}",
			},
		},
	}

	env := internal.Environment{
		Security: internal.Security{Auth: internal.AuthMap{
			"auth-id": {
				Type:         "OAuth2",
				TokenURL:     server.URL,
				GrantType:    "Client Credentials",
				ClientID:     "client-id",
				ClientSecret: "client-secret",
			},
		}},
	}

	err := Auth(&requests, env)

	assert.NoError(t, err)
	assert.Equal(t, "Bearer abc-token", requests[0].Headers["Authorization"])
	assert.Equal(t, "Bearer abc-token", requests[1].Headers["Authorization"])
	assert.Equal(t, int32(1), atomic.LoadInt32(&tokenRequests))
}

func TestGetToken_ReturnsErrorForUnsupportedGrantType(t *testing.T) {
	auth := internal.AuthConfig{
		Type:      "OAuth2",
		GrantType: "Unsupported",
		ClientID:  "client-id",
	}

	token, err := GetToken(auth)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "unsupported grant type")
}

func TestGetToken_ReturnsErrorForUnsupportedAuthType(t *testing.T) {
	auth := internal.AuthConfig{
		Type:      "Basic",
		GrantType: "Client Credentials",
		ClientID:  "client-id",
	}

	token, err := GetToken(auth)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "unsupported auth type")
}
