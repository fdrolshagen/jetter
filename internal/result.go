package internal

import "time"

type Result struct {
	Executions []Execution
	AnyError   bool
}

type Execution struct {
	Responses []Response
	AnyError  bool
}

type Response struct {
	Index    int
	Name     string
	Status   int
	Duration time.Duration
	Error    error
}
