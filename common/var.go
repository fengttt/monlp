package common

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/fengttt/gcl"
)

var (
	// WorkingDir of mochat.   Note that his is NOT the current
	// working directory of the process.
	WorkingDir string
	// Sql database
	SqlDriver string
	// Level of verbosity 0-3
	Verbose int
	// LLM Model
	LLMModel string
	// LLM Temperature
	LLMTemp float64
)

func decideFlagValue(envName, flagValue, dflt string) string {
	if flagValue != "" {
		return flagValue
	}

	v := os.Getenv(envName)
	if v != "" {
		return v
	}

	return dflt
}

func ParseFlags(args []string) {
	fWD := flag.String("d", "", "Working directory")
	sqlDr := flag.String("db", "", "Sql driver")
	llm := flag.String("llm", "", "LLM model")
	llmTemp := flag.Float64("temp", 0.0, "LLM temperature")

	v1 := flag.Bool("v", false, "Verbose")
	v2 := flag.Bool("vv", false, "Verbose2")
	v3 := flag.Bool("vvv", false, "Verbose3")

	flag.CommandLine.Parse(args)

	WorkingDir = decideFlagValue("MOCHAT_WORKING_DIR", *fWD, "")
	SqlDriver = decideFlagValue("MOCHAT_SQL_DRIVER", *sqlDr, "mysql")
	LLMModel = decideFlagValue("MOCHAT_LLM_MODEL", *llm, "deepseek-r1:32b")
	LLMTemp = *llmTemp

	if *v1 {
		Verbose = 1
	}
	if *v2 {
		Verbose = 2
	}
	if *v3 {
		Verbose = 3
	}

	setupWorkingDir()
	setupLogger()
}

func setupWorkingDir() {
	if WorkingDir == "" {
		homedir := gcl.Must(os.UserHomeDir())
		WorkingDir = path.Join(homedir, ".mochat")
	}
}

func setupLogger() {
	lf, err := os.OpenFile(path.Join(WorkingDir, "mochat.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic("Failed to open log file: " + err.Error())
	}

	lv := slog.LevelError
	switch Verbose {
	case 0:
		// we log at error level by default, be concise
	case 1:
		lv = slog.LevelWarn
	case 2:
		lv = slog.LevelInfo
	case 3:
		lv = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: lv,
	}
	logger := slog.New(slog.NewJSONHandler(lf, opts))
	slog.SetDefault(logger)
}

func DbConnInfoForTest() (string, string) {
	var driver string
	var connstr string
	switch SqlDriver {
	case "sqlite", "sqlite3", "dslite", "dslite3":
		driver = "sqlite3"
		connstr = path.Join(WorkingDir, "monlp.db")
	default:
		driver = "mysql"
		connstr = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			"dump", "111", "localhost", "6001", "monlp")
	}
	return driver, connstr
}
