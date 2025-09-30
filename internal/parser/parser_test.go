package parser

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestParseHttp_ShouldParseSingleRequest(t *testing.T) {
	content := strings.TrimSpace(
		`
		###
		POST http://localhost:8081/users
		Content-Type: application/json

		{"name": "foobar"}
		`)

	c, err := ParseHttp(strings.NewReader(content))

	assert.Nil(t, err)

	requests := c.Requests
	assert.NotNil(t, requests)
	assert.Len(t, requests, 1)

	req := requests[0]
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "http://localhost:8081/users", req.Url)
	assert.Contains(t, req.Body, "foobar")
	assert.Equal(t, "application/json", req.Headers["Content-Type"])
}

func TestParseHttp_ShouldParseTwoRequest(t *testing.T) {
	content := strings.TrimSpace(
		`
		### Create New User
		POST http://localhost:8081/users
		Content-Type: application/json

		{"name": "foobar"}

		### Get All Users
		GET http://localhost:8081/users
		Content-Type: application/json
		`)

	c, err := ParseHttp(strings.NewReader(content))

	assert.Nil(t, err)

	requests := c.Requests
	assert.NotNil(t, requests)
	assert.Len(t, requests, 2)
}

func TestParseHttp_ShouldGenerateNameWhenNotGiven(t *testing.T) {
	content := strings.TrimSpace(
		`
		###
		GET http://localhost:8081/users
		`)

	c, err := ParseHttp(strings.NewReader(content))

	assert.Nil(t, err)

	requests := c.Requests
	assert.NotNil(t, requests)
	assert.Len(t, requests, 1)

	req := requests[0]
	assert.Equal(t, "Request #1", req.Name)
}

func TestParseHttp_ShouldDefaultToGET(t *testing.T) {
	content := `
		### request
		http://localhost:8081/users
		`

	c, err := ParseHttp(strings.NewReader(content))

	assert.Nil(t, err)

	requests := c.Requests
	req := requests[0]
	assert.Equal(t, "GET", req.Method)
}

func TestParseHttp_ShouldParseError(t *testing.T) {
	content := `
		### request
		GET
		`

	_, err := ParseHttp(strings.NewReader(content))

	assert.NotNil(t, err)
}

func TestParseHttp_ShouldParseGlobalVariable(t *testing.T) {
	content := strings.TrimSpace(
		`
		@ID = 123

		###
		GET http://localhost:8081/users
		`)

	c, err := ParseHttp(strings.NewReader(content))

	assert.Nil(t, err)

	vars := c.Variables

	assert.Len(t, vars, 1)
	assert.Equal(t, "123", vars["ID"])
}

func TestParseHttp_ShouldParseMultipleGlobalVariables(t *testing.T) {
	content := strings.TrimSpace(
		`
		@ID = 123
		@TSID = 0{{$random.hexadecimal(12)}}

		###
		GET http://localhost:8081/users
		`)

	c, err := ParseHttp(strings.NewReader(content))

	assert.Nil(t, err)

	vars := c.Variables

	assert.Len(t, vars, 2)
	assert.Equal(t, "123", vars["ID"])
	assert.Equal(t, "0{{$random.hexadecimal(12)}}", vars["TSID"])
}

func TestParseHttp_ShouldErrorOnParseGlobalVariable(t *testing.T) {
	content := strings.TrimSpace(
		`
		@ID 123

		###
		GET http://localhost:8081/users
		`)

	_, err := ParseHttp(strings.NewReader(content))

	assert.NotNil(t, err)
}
