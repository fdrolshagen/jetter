package executor

import (
	"github.com/fdrolshagen/jetter/internal"
	"net/http"
	"strings"
	"time"
)

func Submit(s internal.Scenario) internal.Result {
	result := internal.Result{Executions: make([]internal.Execution, 0)}

	if s.Once || s.Duration == 0 {
		execution := ExecuteScenario(s)
		result.Executions = append(result.Executions, execution)
		if execution.AnyError {
			result.AnyError = true
		}
		return result
	}

	start := time.Now()
	for s.Duration >= time.Since(start) {
		execution := ExecuteScenario(s)
		result.Executions = append(result.Executions, execution)
		if result.AnyError {
			result.AnyError = true
		}
		time.Sleep(10 * time.Millisecond)
	}

	return result
}

func ExecuteScenario(s internal.Scenario) internal.Execution {
	responses := make([]internal.Response, 0)
	anyError := false
	for index, request := range s.Requests {
		response := ExecuteRequest(request)
		response.Index = index
		responses = append(responses, response)
		if response.Error != nil {
			anyError = true
		}
	}
	return internal.Execution{Responses: responses, AnyError: anyError}
}

func ExecuteRequest(r internal.Request) internal.Response {
	result := internal.Response{Error: nil, Name: r.Name}
	req, err := http.NewRequest(r.Method, r.Url, strings.NewReader(r.Body))
	if err != nil {
		result.Error = err
		return result
	}

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		result.Error = err
		return result
	}
	elapsed := time.Since(start)

	result.Duration = elapsed
	result.Status = resp.StatusCode
	return result
}
