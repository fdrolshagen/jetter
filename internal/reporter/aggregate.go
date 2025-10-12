package reporter

import (
	"github.com/fdrolshagen/jetter/internal"
	"sort"
	"time"
)

type Metrics struct {
	Index       int
	Name        string
	Total       int
	Failed      int
	Fastest     time.Duration
	Slowest     time.Duration
	Average     time.Duration
	Durations   []time.Duration
	StatusCodes map[int]int
}

func Aggregate(result internal.Result) []Metrics {
	m := make(map[int]*Metrics)

	for _, exec := range result.Executions {
		for _, resp := range exec.Responses {
			metric, ok := m[resp.Index]
			if !ok {
				metric = &Metrics{
					Index:       resp.Index,
					Name:        resp.Name,
					StatusCodes: make(map[int]int),
				}
				m[resp.Index] = metric
			}

			metric.Total++
			metric.Durations = append(metric.Durations, resp.Duration)

			// Count HTTP status codes
			if resp.Status > 0 {
				metric.StatusCodes[resp.Status]++
			}

			// Count failures
			if resp.Error != nil || resp.Status >= 400 {
				metric.Failed++
			}
		}
	}

	items := make([]Metrics, 0, len(m))
	for _, mm := range m {
		sort.Slice(mm.Durations, func(i, j int) bool { return mm.Durations[i] < mm.Durations[j] })
		mm.Fastest = mm.Durations[0].Round(time.Millisecond)
		mm.Slowest = mm.Durations[len(mm.Durations)-1].Round(time.Millisecond)
		mm.Average = mean(mm.Durations).Round(time.Millisecond)
		items = append(items, *mm)
	}

	sort.Slice(items, func(i, j int) bool { return items[i].Index < items[j].Index })
	return items
}

func mean(nums []time.Duration) time.Duration {
	if len(nums) == 0 {
		return 0
	}

	var sum time.Duration
	for _, n := range nums {
		sum += n
	}
	return sum / time.Duration(len(nums))
}
