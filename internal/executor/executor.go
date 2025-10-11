package executor

import (
	"bytes"
	"github.com/fdrolshagen/jetter/internal"
	"net/http"
	"sync"
	"time"
)

func Submit(s internal.Scenario) internal.Result {
	if s.Duration == 0 {
		execution := ExecuteScenario(s)
		return internal.Result{
			Executions: []internal.Execution{execution},
			AnyError:   execution.AnyError,
		}
	}

	start := time.Now()
	duration := s.Duration
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
			for time.Since(start) < duration {
				resultsCh <- ExecuteScenario(s)
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var result internal.Result
	for execution := range resultsCh {
		result.AnyError = execution.AnyError
		result.Executions = append(result.Executions, execution)
	}

	return result
}

func ExecuteScenario(s internal.Scenario) internal.Execution {
	requests, err := Evaluate(s.Collection)
	if err != nil {
		return internal.Execution{
			Responses: nil,
			AnyError:  true,
		}
	}

	responses := make([]internal.Response, 0, len(requests))
	anyError := false
	for index, request := range requests {
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
	req, err := http.NewRequest(r.Method, r.Url, bytes.NewBuffer([]byte(r.Body)))
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
	elapsed := time.Since(start)

	result.Duration = elapsed
	result.Status = resp.StatusCode
	return result
}
