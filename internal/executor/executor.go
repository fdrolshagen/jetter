package executor

import (
	"bytes"
	"context"
	"fmt"
	"github.com/fdrolshagen/jetter/internal"
	"github.com/fdrolshagen/jetter/internal/script"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ResponseCallback func(response internal.Response)

// Submit executes the given scenario either once or concurrently for a specified duration.
//
// If the scenario has a duration of zero, a single execution is performed.
// Otherwise, multiple executions are run concurrently according to the scenario's concurrency setting,
// and they continue until the duration elapses or the context is canceled.
//
// The function aggregates the results of all executions and indicates whether any of them encountered an error.
func Submit(s internal.Scenario) internal.Result {
	return SubmitWithResponseCallback(s, nil)
}

func SubmitWithResponseCallback(s internal.Scenario, responseCallback ResponseCallback) internal.Result {
	if s.Duration == 0 {
		execution := executeScenario(context.Background(), s, responseCallback)
		return internal.Result{
			Executions: []internal.Execution{execution},
			AnyError:   execution.AnyError,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Duration)
	defer cancel()

	numWorkers := s.Concurrency
	if numWorkers <= 0 {
		numWorkers = 1
	}

	resultsCh := make(chan internal.Execution, 1000)
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					resultsCh <- executeScenario(ctx, s, responseCallback)
					time.Sleep(10 * time.Millisecond)
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var result internal.Result
	for execution := range resultsCh {
		result.Executions = append(result.Executions, execution)
		if execution.AnyError {
			result.AnyError = true
		}
	}

	return result
}

// ExecuteScenario executes all requests defined by the given scenario within the provided context.
//
// The scenario may include multiple requests, which are executed sequentially in order.
// The provided context `ctx` is used for cancellation and timeout; if the context is done,
// in-progress requests will be interrupted.
//
// The returned Execution summarizes the results of all requests and indicates whether
// any of them encountered an error.
func ExecuteScenario(ctx context.Context, s internal.Scenario) internal.Execution {
	return executeScenario(ctx, s, nil)
}

func executeScenario(ctx context.Context, s internal.Scenario, responseCallback ResponseCallback) internal.Execution {
	vars, err := s.Collection.EvaluateVariables()
	if err != nil {
		return internal.Execution{
			Responses: nil,
			AnyError:  true,
		}
	}

	responses := make([]internal.Response, 0, len(s.Collection.Requests))
	anyError := false
	for index, request := range s.Collection.Requests {
		onResponse := func(response internal.Response) {
			response.Index = index
			if response.Error != nil {
				anyError = true
			}
			if responseCallback != nil {
				responseCallback(response)
			}
			responses = append(responses, response)
		}

		executeRequestWithLoop(ctx, request, vars, onResponse)
	}

	return internal.Execution{Responses: responses, AnyError: anyError}
}

func executeRequestWithLoop(ctx context.Context, request internal.Request, vars map[string]string, onResponse func(internal.Response)) []internal.Response {
	responses := make([]internal.Response, 0, 1)
	whileEnabled := strings.TrimSpace(request.JetterWhile) != ""

	maxIterations := request.JetterMaxIterations
	if maxIterations <= 0 {
		maxIterations = 1
	}

	whileCondition := replaceVariablesInString(request.JetterWhile, vars)
	iteration := 1
	for {
		response := executeRequestAndPostScript(ctx, request, vars)
		if response.Error != nil {
			responses = append(responses, response)
			onResponse(response)
			return responses
		}

		if !whileEnabled {
			responses = append(responses, response)
			onResponse(response)
			return responses
		}

		shouldContinue, err := script.EvaluateWhileCondition(whileCondition, response, vars)
		if err != nil {
			response.Error = err
			responses = append(responses, response)
			onResponse(response)
			return responses
		}
		if !shouldContinue {
			responses = append(responses, response)
			onResponse(response)
			return responses
		}

		if iteration >= maxIterations {
			if shouldFailOnTimeout(request.JetterOnTimeout) {
				response.Error = fmt.Errorf("jetter while condition still true after %d iterations", iteration)
			}
			responses = append(responses, response)
			onResponse(response)
			return responses
		}

		responses = append(responses, response)
		onResponse(response)

		if err := sleepWithContext(ctx, request.JetterSleep); err != nil {
			responses[len(responses)-1].Error = err
			return responses
		}

		iteration++
	}
}

func executeRequestAndPostScript(ctx context.Context, request internal.Request, vars map[string]string) internal.Response {
	evaluatedRequest := request
	evaluatedRequest.Url = replaceVariablesInString(request.Url, vars)
	evaluatedRequest.Body = replaceVariablesInString(request.Body, vars)

	evaluatedHeaders := make(map[string]string, len(request.Headers))
	for key, value := range request.Headers {
		evaluatedHeaders[key] = replaceVariablesInString(value, vars)
	}
	evaluatedRequest.Headers = evaluatedHeaders

	response := ExecuteRequest(ctx, evaluatedRequest)
	if response.Error != nil {
		return response
	}

	postScript := replaceVariablesInString(request.PostScript, vars)
	if postScript != "" {
		if err := script.ExecutePostScript(postScript, response, vars); err != nil {
			response.Error = err
		}
	}

	return response
}

func sleepWithContext(ctx context.Context, sleep time.Duration) error {
	if sleep <= 0 {
		return nil
	}

	timer := time.NewTimer(sleep)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func shouldFailOnTimeout(onTimeout string) bool {
	if strings.ToLower(strings.TrimSpace(onTimeout)) == "continue" {
		return false
	}
	return true
}

// ExecuteRequest performs a single HTTP request described by the given internal.Request.
//
// The request uses the provided context `ctx`, which may include cancellation or a timeout.
// If `ctx` has no deadline, a default timeout (e.g., 5 seconds) is applied to prevent
// the request from hanging indefinitely.
//
// The returned internal.Response contains the HTTP status code, the elapsed duration of
// the request, and any error encountered during creation or execution.
func ExecuteRequest(ctx context.Context, r internal.Request) internal.Response {
	ctx, cancel := withDefaultTimeout(ctx, 5*time.Second)
	defer cancel()

	result := internal.Response{Error: nil, Name: r.Name}
	req, err := http.NewRequestWithContext(ctx, r.Method, r.Url, bytes.NewBuffer([]byte(r.Body)))
	if err != nil {
		result.Error = err
		return result
	}

	for key, value := range r.Headers {
		req.Header.Set(key, value)
	}

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		result.Error = err
		return result
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = err
		return result
	}
	elapsed := time.Since(start)

	headers := make(map[string]string, len(resp.Header))
	for key, values := range resp.Header {
		if len(values) == 0 {
			continue
		}
		headers[key] = values[0]
	}

	result.Duration = elapsed
	result.Status = resp.StatusCode
	result.Headers = headers
	result.Body = string(body)
	return result
}

// withDefaultTimeout returns a context with the given timeout
// if the original context has no deadline set.
func withDefaultTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); !ok {
		return context.WithTimeout(ctx, timeout)
	}
	return ctx, func() {}
}
