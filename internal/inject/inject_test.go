package inject

import (
	"github.com/fdrolshagen/jetter/internal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInject_MergesEnvironmentVariables(t *testing.T) {
	collection := &internal.Collection{
		Variables: map[string]string{"LOCAL": "keep"},
		Requests:  []internal.Request{{Method: "GET", Url: "http://localhost"}},
	}

	env := internal.Environment{
		Variables: map[string]string{"URL": "http://api.local", "LOCAL": "override"},
	}

	err := Inject(collection, env)

	assert.NoError(t, err)
	assert.Equal(t, "keep", collection.Variables["LOCAL"])
	assert.Equal(t, "http://api.local", collection.Variables["URL"])
}

func TestInject_ReturnsAuthErrorAndStillMergesVariables(t *testing.T) {
	collection := &internal.Collection{
		Variables: map[string]string{},
		Requests: []internal.Request{
			{
				Method: "GET",
				Url:    "http://localhost",
				Headers: map[string]string{
					"Authorization": "Bearer {{$auth.token(auth-id)}}",
				},
			},
		},
	}

	env := internal.Environment{
		Variables: map[string]string{"URL": "http://api.local"},
	}

	err := Inject(collection, env)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid auth token placeholder")
	assert.Equal(t, "http://api.local", collection.Variables["URL"])
}
