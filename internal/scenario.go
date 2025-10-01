package internal

import "time"

type Scenario struct {
	Collection  *Collection
	Concurrency int
	Duration    time.Duration
}
