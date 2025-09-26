package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/fdrolshagen/jetter/internal"
	"github.com/fdrolshagen/jetter/internal/executor"
	"github.com/fdrolshagen/jetter/internal/parser"
	"github.com/fdrolshagen/jetter/internal/reporter"
	"os"
	"time"
)

func Execute() {
	PrintBanner()

	pending := "⏳"
	success := "✔"

	msg := "Initializing Scenario..."
	fmt.Printf("%s %s", pending, msg)
	requests, err := parser.ParseHttpFile("./examples/example.http")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	s := internal.Scenario{
		Once:     false,
		Requests: requests,
		Duration: 1 * time.Second,
	}
	fmt.Printf("\r%s %s\n", color.GreenString(success), msg)

	msg = "Running Scenario..."
	fmt.Printf("%s %s", pending, msg)
	result := executor.Submit(s)
	fmt.Printf("\r%s %s\n\n", color.GreenString(success), msg)

	reporter.Report(result)
}
