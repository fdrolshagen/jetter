package script

import (
	"encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"github.com/fdrolshagen/jetter/internal"
	"io"
	"os"
	"strings"
)

var consoleWriter io.Writer = os.Stdout

func ExecutePostScript(postScript string, response internal.Response, variables map[string]string) error {
	if strings.TrimSpace(postScript) == "" {
		return nil
	}

	runtime := goja.New()

	body := any(response.Body)
	var parsedBody any
	if err := json.Unmarshal([]byte(response.Body), &parsedBody); err == nil {
		body = parsedBody
	}

	responseData := map[string]any{
		"status":  response.Status,
		"headers": response.Headers,
		"body":    body,
	}

	clientData := map[string]any{
		"global": map[string]any{
			"set": func(key string, value any) {
				variables[key] = fmt.Sprint(value)
			},
			"get": func(key string) any {
				value, exists := variables[key]
				if !exists {
					return nil
				}
				return value
			},
		},
	}

	consoleData := map[string]any{
		"log": func(args ...any) {
			values := make([]string, 0, len(args))
			for _, arg := range args {
				values = append(values, stringifyConsoleArg(arg))
			}
			_, _ = fmt.Fprintln(consoleWriter, strings.Join(values, " "))
		},
	}

	if err := runtime.Set("response", responseData); err != nil {
		return fmt.Errorf("failed to prepare response context: %w", err)
	}
	if err := runtime.Set("client", clientData); err != nil {
		return fmt.Errorf("failed to prepare client context: %w", err)
	}
	if err := runtime.Set("console", consoleData); err != nil {
		return fmt.Errorf("failed to prepare console context: %w", err)
	}

	if _, err := runtime.RunString(postScript); err != nil {
		return fmt.Errorf("post-script execution failed: %w", err)
	}

	return nil
}

func stringifyConsoleArg(value any) string {
	if value == nil {
		return "null"
	}
	if str, ok := value.(string); ok {
		return str
	}

	jsonValue, err := json.Marshal(value)
	if err == nil {
		return string(jsonValue)
	}

	return fmt.Sprint(value)
}
