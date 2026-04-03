package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/fdrolshagen/jetter/internal"
	"github.com/fdrolshagen/jetter/internal/executor"
	"github.com/fdrolshagen/jetter/internal/inject"
	"github.com/fdrolshagen/jetter/internal/logger"
	"github.com/fdrolshagen/jetter/internal/parser"
	"github.com/fdrolshagen/jetter/internal/reporter"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var (
	duration    time.Duration
	concurrency int
	file        string
	envPath     string
	logPath     string
	showVersion bool
)

const (
	pendingIcon = "⏳"
	successIcon = "✔"
)

func Execute() {
	var exitCode int
	rootCmd := &cobra.Command{
		Use:   "jetter",
		Short: "Jetter – a load test tool",
		Long:  "Jetter runs load tests based on .http scenario files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			PrintBanner()
			exitCode = run()
			return nil
		},
	}

	rootCmd.PersistentFlags().BoolVar(&showVersion, "version", false, "Print version and exit")
	rootCmd.Flags().DurationVarP(&duration, "duration", "d", 0,
		"How long should the load test run (accepts duration format, e.g. 30s, 1m)")
	rootCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 1, "Number of concurrent workers")
	rootCmd.Flags().StringVarP(&file, "file", "f", "", "Path to the .http file")
	rootCmd.Flags().StringVarP(&envPath, "env", "e", "", "Path to the environment file")
	rootCmd.Flags().StringVar(&logPath, "log", "", "Path to newline-delimited JSON request logs")
	rootCmd.MarkFlagRequired("file")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if showVersion {
			fmt.Println(internal.Version)
			os.Exit(0)
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

func run() int {
	msg := "Parsing .http file..."
	fmt.Printf("%s %s", pendingIcon, msg)
	collection, err := parser.ParseHttpFile(file)
	if err != nil {
		PrintError(err)
		os.Exit(1)
	}
	fmt.Printf("\r%s %s\n", color.GreenString(successIcon), msg)

	if envPath != "" {
		err = handleEnvInjection(envPath, &collection)
		if err != nil {
			PrintError(err)
			os.Exit(1)
		}
	}

	s := internal.Scenario{
		Concurrency: concurrency,
		Collection:  &collection,
		Duration:    duration,
	}

	msg = "Running Scenario..."
	fmt.Printf("%s %s\n", pendingIcon, msg)

	var requestLogger *logger.RequestLogger
	if logPath != "" {
		requestLogger, err = logger.NewRequestLogger(logPath)
		if err != nil {
			PrintError(fmt.Errorf("failed to initialize logger: %w", err))
			os.Exit(1)
		}
		defer requestLogger.Close()
	}

	liveReporter := reporter.NewLiveReporter(200 * time.Millisecond)
	liveReporter.Start()
	callback := func(response internal.Response) {
		liveReporter.AddResponse(response)
		if requestLogger != nil {
			requestLogger.Log(response)
		}
	}
	result := executor.SubmitWithResponseCallback(s, callback)
	liveReporter.Stop()
	fmt.Printf("%s %s\n\n", color.GreenString(successIcon), msg)

	if requestLogger != nil {
		if err := requestLogger.Err(); err != nil {
			PrintError(fmt.Errorf("failed to write request log: %w", err))
			result.AnyError = true
		}
	}

	if !liveReporter.Enabled() {
		reporter.Report(result)
	}
	return map[bool]int{true: 1, false: 0}[result.AnyError]
}

func PrintError(err error) {
	if err != nil {
		fmt.Printf("\n\n❌ Error: %s\n", err.Error())
	}
}

func handleEnvInjection(envPath string, collection *internal.Collection) error {
	msg := "Reading Environment..."
	fmt.Printf("%s %s", pendingIcon, msg)
	env, err := parser.ParseEnv(envPath)
	if err != nil {
		return err
	}
	fmt.Printf("\r%s %s\n", color.GreenString(successIcon), msg)

	msg = "Injecting Variables..."
	fmt.Printf("%s %s", pendingIcon, msg)
	err = inject.Inject(collection, env)
	if err != nil {
		return err
	}
	fmt.Printf("\r%s %s\n", color.GreenString(successIcon), msg)

	return nil
}
