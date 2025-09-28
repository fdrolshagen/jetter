package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/fdrolshagen/jetter/internal"
	"github.com/fdrolshagen/jetter/internal/executor"
	"github.com/fdrolshagen/jetter/internal/inject"
	"github.com/fdrolshagen/jetter/internal/parser"
	"github.com/fdrolshagen/jetter/internal/reporter"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var (
	duration time.Duration
	once     bool
	file     string
	envPath  string
)

func Execute() {
	PrintBanner()

	var exitCode int
	rootCmd := &cobra.Command{
		Use:   "jetter",
		Short: "Jetter – a load test tool",
		Long:  "Jetter runs load tests based on .http scenario files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			exitCode = run()
			return nil
		},
	}

	rootCmd.Flags().DurationVarP(&duration, "duration", "d", 0,
		"How long should the load test run (accepts duration format, e.g. 30s, 1m)")
	rootCmd.Flags().BoolVar(&once, "once", false, "run the scenario exactly once (ignores concurrency and duration)")
	rootCmd.Flags().StringVarP(&file, "file", "f", "", "Path to the .http file")
	rootCmd.Flags().StringVarP(&envPath, "env", "e", "", "Path to the environment file")
	rootCmd.MarkFlagRequired("file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

func run() int {
	pending := "⏳"
	success := "✔"

	msg := "Parsing Requests..."
	fmt.Printf("%s %s", pending, msg)
	requests, err := parser.ParseHttpFile(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("\r%s %s\n", color.GreenString(success), msg)

	msg = "Reading Environment..."
	fmt.Printf("%s %s", pending, msg)
	env, err := parser.ParseEnv(envPath)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Printf("\r%s %s\n", color.GreenString(success), msg)

	msg = "Injecting Variables..."
	fmt.Printf("%s %s", pending, msg)
	err = inject.Inject(&requests, env)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Printf("\r%s %s\n", color.GreenString(success), msg)

	s := internal.Scenario{
		Once:     once,
		Requests: requests,
		Duration: duration,
	}

	msg = "Running Scenario..."
	fmt.Printf("%s %s", pending, msg)
	result := executor.Submit(s)
	fmt.Printf("\r%s %s\n\n", color.GreenString(success), msg)

	reporter.Report(result)
	return map[bool]int{true: 1, false: 0}[result.AnyError]
}
