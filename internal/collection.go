package internal

import (
	"fmt"
	"github.com/fdrolshagen/jetter/internal/random"
	"regexp"
)

// Request represents a single HTTP request definition within a jetter scenario.
// It defines all necessary details for execution, including the method, target URL,
// optional headers, and request body content.
type Request struct {
	Name    string
	Method  string
	Url     string
	Headers map[string]string
	Body    string
}

// Collection represents a reusable group of HTTP requests that make up
// a scenario to be executed by jetter. It may also include variable definitions
// that can be referenced within individual requests.
type Collection struct {
	Requests  []Request
	Variables map[string]string
}

var funcRegex = regexp.MustCompile(`\{\{\s*\$([a-zA-Z0-9_]+)\.([a-zA-Z0-9_]+)\((.*?)\)\s*}}`)

func (c *Collection) EvaluateVariables() (map[string]string, error) {
	resolved := make(map[string]string, len(c.Variables))
	for k, v := range c.Variables {
		r, err := replaceFunctions(v, k)
		if err != nil {
			return nil, fmt.Errorf("error in variable '%s': %w", k, err)
		}
		resolved[k] = r
	}
	return resolved, nil
}

func replaceFunctions(input, varName string) (string, error) {
	result := ""
	lastIndex := 0

	matches := funcRegex.FindAllStringSubmatchIndex(input, -1)
	for _, match := range matches {
		start, end := match[0], match[1]
		nsStart, nsEnd := match[2], match[3]
		funcStart, funcEnd := match[4], match[5]
		argStart, argEnd := match[6], match[7]

		result += input[lastIndex:start]

		namespace := input[nsStart:nsEnd]
		funcName := input[funcStart:funcEnd]
		arg := input[argStart:argEnd]

		var out string
		var err error

		switch namespace {
		case "random":
			out, err = random.Execute(funcName, arg)
		default:
			return "", fmt.Errorf("error in variable '%s': unsupported namespace '%s'", varName, namespace)
		}

		if err != nil {
			return "", fmt.Errorf("error in variable '%s': %v", varName, err)
		}

		result += out
		lastIndex = end
	}

	result += input[lastIndex:]
	return result, nil
}

func (c *Collection) MergeEnvironmentVariables(env Environment) {
	if c.Variables == nil {
		c.Variables = make(map[string]string)
	}
	for k, v := range env.Variables {
		// collection variables take precedence over environment variables
		if _, exists := c.Variables[k]; !exists {
			c.Variables[k] = v
		}
	}
}
