package executor

import (
	"bytes"
	"context"
	"github.com/fdrolshagen/jetter/internal"
	"net/http"
	"sync"
	"time"
)

func Submit(s internal.Scenario) internal.Result {
	if s.Duration == 0 {
		execution := ExecuteScenario(context.Background(), s)
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
					resultsCh <- ExecuteScenario(ctx, s)
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

func ExecuteScenario(ctx context.Context, s internal.Scenario) internal.Execution {
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
		response := ExecuteRequest(ctx, request)
		response.Index = index
		responses = append(responses, response)
		if response.Error != nil {
			anyError = true
		}
	}

	return internal.Execution{Responses: responses, AnyError: anyError}
}

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
	elapsed := time.Since(start)

	result.Duration = elapsed
	result.Status = resp.StatusCode
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
