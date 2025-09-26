package internal

import "time"

type Scenario struct {
	Once        bool
	Requests    []Request
	Concurrency int
	Duration    time.Duration
}
