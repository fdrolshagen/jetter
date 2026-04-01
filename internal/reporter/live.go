package reporter

import (
	"fmt"
	"github.com/fdrolshagen/jetter/internal"
	"os"
	"sync"
	"time"
)

type LiveReporter struct {
	mu       sync.Mutex
	renderMu sync.Mutex
	result   internal.Result
	enabled  bool
}

func NewLiveReporter(_ time.Duration) *LiveReporter {
	return &LiveReporter{
		enabled: SupportsLiveOutput(),
	}
}

func SupportsLiveOutput() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	if (info.Mode() & os.ModeCharDevice) == 0 {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return true
}

func (r *LiveReporter) Enabled() bool {
	return r.enabled
}

func (r *LiveReporter) Start() {
}

func (r *LiveReporter) Stop() {
}

func (r *LiveReporter) AddResponse(response internal.Response) {
	if !r.enabled {
		return
	}

	r.mu.Lock()

	exec := internal.Execution{Responses: []internal.Response{response}}
	if response.Error != nil {
		exec.AnyError = true
		r.result.AnyError = true
	}
	r.result.Executions = append(r.result.Executions, exec)
	r.mu.Unlock()

	r.render()
}

func (r *LiveReporter) render() {
	r.renderMu.Lock()
	defer r.renderMu.Unlock()

	r.mu.Lock()
	result := cloneResult(r.result)
	r.mu.Unlock()

	metrics := Aggregate(result)
	fmt.Print("\033[H\033[2J")
	fmt.Println("Live request metrics")
	if len(metrics) == 0 {
		fmt.Println("Waiting for first request to complete...")
		return
	}
	_ = TableReport(metrics)
}

func cloneResult(result internal.Result) internal.Result {
	clone := internal.Result{AnyError: result.AnyError}
	clone.Executions = make([]internal.Execution, len(result.Executions))
	for i, execution := range result.Executions {
		responses := make([]internal.Response, len(execution.Responses))
		copy(responses, execution.Responses)
		clone.Executions[i] = internal.Execution{
			Responses: responses,
			AnyError:  execution.AnyError,
		}
	}
	return clone
}
