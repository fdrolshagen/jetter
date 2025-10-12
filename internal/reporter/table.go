package reporter

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"os"
	"sort"
	"strings"
	"time"
)

func TableReport(metrics []Metrics) error {
	table := configureTableWriter()

	for _, m := range metrics {

		table.Append([]string{
			m.Name,
			fmt.Sprintf("%d", m.Total),
			colorDuration(m.Fastest, m.Fastest, m.Slowest),
			colorDuration(m.Slowest, m.Fastest, m.Slowest),
			colorMean(m.Average, m.Fastest, m.Slowest),
			formatTotalFailed(m.Failed),
			formatStatusCodes(m.StatusCodes),
		})
	}

	table.Render()
	return nil
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

func formatTotalFailed(countFailed int) string {
	if countFailed > 0 {
		return color.RedString("%d", countFailed)
	}
	return color.GreenString("0")
}

func formatStatusCodes(codes map[int]int) string {
	if len(codes) == 0 {
		return "-"
	}

	keys := make([]int, 0, len(codes))
	for k := range codes {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	parts := make([]string, 0, len(keys))
	for _, code := range keys {
		count := codes[code]
		var part string

		switch {
		case code >= 200 && code < 300:
			part = color.GreenString("%d × %d", count, code)
		case code >= 400 && code < 500:
			part = color.YellowString("%d × %d", count, code)
		case code >= 500:
			part = color.RedString("%d × %d", count, code)
		default:
			part = fmt.Sprintf("%d × %d", count, code)
		}

		parts = append(parts, part)
	}

	return strings.Join(parts, "   ")
}

func configureTableWriter() *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Total", "Fastest", "Longest", "Mean", "Failed", "Status Codes"})
	table.SetBorders(tablewriter.Border{Left: false, Top: true, Right: false, Bottom: true})
	table.SetAutoWrapText(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_CENTER,
	})
	table.SetHeaderLine(true)
	table.SetRowLine(true)
	table.SetCenterSeparator("│")
	table.SetColumnSeparator("│")
	table.SetRowSeparator("─")
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiWhiteColor},
	)
	return table
}
