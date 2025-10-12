package internal

import "time"

// Result represents the overall outcome of a scenario run within jetter.
// It aggregates all Executions performed as part of the scenario and
// indicates whether any of them encountered an error.
type Result struct {
	Executions []Execution
	AnyError   bool
}

// Execution represents the result of a single scenario execution,
// typically corresponding to one logical request sequence.
// It captures all individual Responses and whether any of them failed.
type Execution struct {
	Responses []Response
	AnyError  bool
}

// Response represents the outcome of a single request within a scenario execution.
// It contains metadata such as the request name, response status,
// execution duration, and any associated error.
type Response struct {
	Index    int
	Name     string
	Status   int
	Duration time.Duration
	Error    error
}
