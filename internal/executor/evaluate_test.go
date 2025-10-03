package executor

import (
	"github.com/fdrolshagen/jetter/internal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvaluate_ReplacesVariablesInUrlBodyAndHeaders(t *testing.T) {
	c := &internal.Collection{
		Variables: map[string]string{
			"ID":    "123",
			"TOKEN": "abc",
		},
		Requests: []internal.Request{
			{
				Method: "GET",
				Url:    "http://localhost/users/{{ID}}",
				Body:   "token={{TOKEN}}",
				Headers: map[string]string{
					"Authorization": "Bearer {{TOKEN}}",
				},
			},
		},
	}

	requests, err := Evaluate(c)
	assert.Nil(t, err)
	assert.Len(t, requests, 1)
	assert.Equal(t, "http://localhost/users/123", requests[0].Url)
	assert.Equal(t, "token=abc", requests[0].Body)
	assert.Equal(t, "Bearer abc", requests[0].Headers["Authorization"])
}

func TestEvaluate_MultipleRequestsAndNoVariables(t *testing.T) {
	c := &internal.Collection{
		Variables: map[string]string{},
		Requests: []internal.Request{
			{
				Method:  "GET",
				Url:     "http://localhost/users",
				Body:    "",
				Headers: map[string]string{},
			},
			{
				Method:  "POST",
				Url:     "http://localhost/users",
				Body:    "{\"name\":\"foo\"}",
				Headers: map[string]string{"Content-Type": "application/json"},
			},
		},
	}

	requests, err := Evaluate(c)
	assert.Nil(t, err)
	assert.Len(t, requests, 2)
	assert.Equal(t, "http://localhost/users", requests[0].Url)
	assert.Equal(t, "{\"name\":\"foo\"}", requests[1].Body)
	assert.Equal(t, "application/json", requests[1].Headers["Content-Type"])
}

func TestEvaluate_EmptyCollection(t *testing.T) {
	c := &internal.Collection{
		Variables: map[string]string{},
		Requests:  []internal.Request{},
	}

	requests, err := Evaluate(c)
	assert.Nil(t, err)
	assert.Len(t, requests, 0)
}
