package internal

import "time"

// Scenario represents an executable load or functional test definition within jetter.
// It specifies which request collection to run, how many executions to perform concurrently,
// and for how long the scenario should be executed.
type Scenario struct {
	Collection  *Collection
	Concurrency int
	Duration    time.Duration
}
