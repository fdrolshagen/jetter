package reporter

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/fdrolshagen/jetter/internal"
	"github.com/olekukonko/tablewriter"
	"sort"
	"time"
)

type metrics struct {
	index     int
	name      string
	durations []time.Duration
	failed    int
	total     int
}

type item struct {
	name string
	*metrics
}

func Report(r internal.Result) {
	table := ConfigureTableWriter()
	items := aggregate(r)
	fillTable(items, table)
	table.Render()
}

func aggregate(result internal.Result) []*metrics {
	m := make(map[int]*metrics)

	for _, exec := range result.Executions {
		for _, resp := range exec.Responses {
			mm, ok := m[resp.Index]
			if !ok {
				mm = &metrics{
					index: resp.Index,
					name:  resp.Name,
				}
				m[resp.Index] = mm
			}
			mm.total++
			mm.durations = append(mm.durations, resp.Duration)
			if resp.Error != nil || resp.Status >= 400 {
				mm.failed++
			}
		}
	}

	items := make([]*metrics, 0, len(m))
	for _, mm := range m {
		items = append(items, mm)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].index < items[j].index
	})
	return items
}

func fillTable(items []*metrics, table *tablewriter.Table) {

	for _, it := range items {
		sort.Slice(it.durations, func(i, j int) bool { return it.durations[i] < it.durations[j] })
		fastest := it.durations[0].Round(time.Millisecond)
		longest := it.durations[len(it.durations)-1].Round(time.Millisecond)
		avg := mean(it.durations).Round(time.Millisecond)

		failedStr := ""
		if it.failed == 0 {
			failedStr = color.GreenString("✅ 0")
		} else {
			failedStr = color.RedString("❌ %d", it.failed)
		}

		table.Append([]string{
			it.name,
			fmt.Sprintf("%d", it.total),
			colorDuration(fastest, fastest, longest),
			colorDuration(longest, fastest, longest),
			colorMean(avg, fastest, longest),
			failedStr,
		})
	}

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

func colorDuration(d, fastest, longest time.Duration) string {
	durStr := d.String()
	switch {
	case d == fastest:
		return color.GreenString("⬆ %s", durStr)
	case d == longest:
		return color.RedString("⬇ %s", durStr)
	default:
		return durStr
	}
}

func colorMean(mean, fastest, longest time.Duration) string {
	if longest == fastest {
		return color.GreenString("⬆ %s", mean.String())
	}

	ratio := float64(mean-fastest) / float64(longest-fastest)
	switch {
	case ratio < 0.33:
		return color.GreenString("⬆ %s", mean.String())
	case ratio < 0.66:
		return color.YellowString("→ %s", mean.String())
	default:
		return color.RedString("⬇ %s", mean.String())
	}
}
