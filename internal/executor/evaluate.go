package executor

import (
	"github.com/fdrolshagen/jetter/internal"
	"strings"
)

func Evaluate(c *internal.Collection) ([]internal.Request, error) {
	vars, err := c.EvaluateVariables()
	if err != nil {
		return nil, err
	}

	requests := make([]internal.Request, 0, len(c.Requests))
	for _, req := range c.Requests {
		newReq := req
		newReq.Url = replaceVariablesInString(newReq.Url, vars)
		newReq.Body = replaceVariablesInString(newReq.Body, vars)
		newHeaders := make(map[string]string, len(newReq.Headers))
		for hk, hv := range newReq.Headers {
			newHeaders[hk] = replaceVariablesInString(hv, vars)
		}
		newReq.Headers = newHeaders
		requests = append(requests, newReq)
	}

	return requests, nil
}

func replaceVariablesInString(input string, vars map[string]string) string {
	result := input
	for k, v := range vars {
		placeholder := "{{" + k + "}}"
		result = strings.ReplaceAll(result, placeholder, v)
	}
	return result
}
