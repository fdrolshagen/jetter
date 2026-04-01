package script

import (
	"bytes"
	"github.com/fdrolshagen/jetter/internal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExecutePostScript_SetsGlobalVariableFromJSONResponseBody(t *testing.T) {
	variables := map[string]string{}
	response := internal.Response{
		Status: 200,
		Body:   `{"token":"abc123"}`,
	}

	err := ExecutePostScript(`client.global.set("TOKEN", response.body.token)`, response, variables)

	assert.NoError(t, err)
	assert.Equal(t, "abc123", variables["TOKEN"])
}

func TestExecutePostScript_ReturnsErrorForInvalidJavaScript(t *testing.T) {
	variables := map[string]string{}
	response := internal.Response{Status: 200, Body: `{}`}

	err := ExecutePostScript(`client.global.set("TOKEN"`, response, variables)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "post-script execution failed")
}

func TestExecutePostScript_SupportsConsoleLog(t *testing.T) {
	buffer := bytes.Buffer{}
	previousWriter := consoleWriter
	consoleWriter = &buffer
	defer func() {
		consoleWriter = previousWriter
	}()

	variables := map[string]string{}
	response := internal.Response{Status: 200, Body: `{"ok":true}`}

	err := ExecutePostScript(`console.log("status", response.status, response.body.ok)`, response, variables)

	assert.NoError(t, err)
	assert.Equal(t, "status 200 true\n", buffer.String())
}
