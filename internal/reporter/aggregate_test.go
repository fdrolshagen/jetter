package reporter

import (
	"errors"
	"github.com/fdrolshagen/jetter/internal"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAggregate(t *testing.T) {
	t.Run("single response", func(t *testing.T) {
		result := internal.Result{
			Executions: []internal.Execution{
				{
					Responses: []internal.Response{
						{
							Index:    0,
							Name:     "GET /users",
							Status:   200,
							Duration: 50 * time.Millisecond,
						},
					},
				},
			},
		}

		metrics := Aggregate(result)
		assert.Len(t, metrics, 1)
		m := metrics[0]
		assert.Equal(t, "GET /users", m.Name)
		assert.Equal(t, 1, m.Total)
		assert.Equal(t, 0, m.Failed)
		assert.Equal(t, 50*time.Millisecond, m.Fastest)
		assert.Equal(t, 50*time.Millisecond, m.Slowest)
		assert.Equal(t, 50*time.Millisecond, m.Average)
		assert.Equal(t, map[int]int{200: 1}, m.StatusCodes)
	})

	t.Run("multiple responses with mixed results", func(t *testing.T) {
		result := internal.Result{
			Executions: []internal.Execution{
				{
					Responses: []internal.Response{
						{
							Index:    1,
							Name:     "POST /login",
							Status:   200,
							Duration: 40 * time.Millisecond,
						},
						{
							Index:    1,
							Name:     "POST /login",
							Status:   500,
							Duration: 120 * time.Millisecond,
							Error:    errors.New("internal error"),
						},
						{
							Index:    1,
							Name:     "POST /login",
							Status:   404,
							Duration: 80 * time.Millisecond,
						},
					},
				},
			},
		}

		metrics := Aggregate(result)
		assert.Len(t, metrics, 1)
		m := metrics[0]

		assert.Equal(t, "POST /login", m.Name)
		assert.Equal(t, 3, m.Total)
		assert.Equal(t, 2, m.Failed)
		assert.Equal(t, 40*time.Millisecond, m.Fastest)
		assert.Equal(t, 120*time.Millisecond, m.Slowest)
		assert.Equal(t, 80*time.Millisecond, m.Average)
		assert.Equal(t, map[int]int{200: 1, 404: 1, 500: 1}, m.StatusCodes)
	})

	t.Run("handles multiple requests correctly", func(t *testing.T) {
		result := internal.Result{
			Executions: []internal.Execution{
				{
					Responses: []internal.Response{
						{Index: 0, Name: "GET /ping", Status: 200, Duration: 10 * time.Millisecond},
						{Index: 1, Name: "GET /health", Status: 503, Duration: 200 * time.Millisecond},
					},
				},
			},
		}

		metrics := Aggregate(result)
		assert.Len(t, metrics, 2)

		assert.Equal(t, "GET /ping", metrics[0].Name)
		assert.Equal(t, "GET /health", metrics[1].Name)
		assert.Equal(t, map[int]int{200: 1}, metrics[0].StatusCodes)
		assert.Equal(t, map[int]int{503: 1}, metrics[1].StatusCodes)
	})
}
