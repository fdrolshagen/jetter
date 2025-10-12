package reporter

import (
	"github.com/fdrolshagen/jetter/internal"
)

func Report(r internal.Result) {
	metrics := Aggregate(r)
	err := TableReport(metrics)
	if err != nil {
		return
	}
}
