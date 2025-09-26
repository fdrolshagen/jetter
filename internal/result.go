package internal

import "time"

type Result struct {
	Executions []Execution
}

type Execution struct {
	Responses []Response
}

type Response struct {
	Index    int
	Name     string
	Status   int
	Duration time.Duration
	Error    error
}
