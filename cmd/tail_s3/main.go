package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
)

var (
	version   string
	buildDate string
)

func main() {
	var (
		showHelp    bool
		showVersion bool
	)
	flag.BoolVar(&showHelp, "help", false, "show help")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.Parse()

	if showVersion {
		printVersion()
		return
	} else if showHelp {
		printUsage(0)
		return
	}

	if flag.NArg() <= 0 {
		fmt.Println("config file is required")
		printUsage(1)
		return
	}

	config, err := LoadConfig(flag.Arg(0))
	if err != nil {
		printError(err.Error())
		return
	}

	logger := &logger{log.New(os.Stdout, "", log.LstdFlags)}

	worker := newInTailOutS3(*config, logger)
	if err := worker.run(); err != nil {
		printError(err.Error())
		return
	}
	os.Exit(0)
}

func printUsage(code int) {
	fmt.Println("Usage of", commandName)
	fmt.Println("")
	fmt.Println(" ", commandName, "[options] INPUT_FILE")
	fmt.Println("")
	flag.PrintDefaults()
	os.Exit(code)
}

func printVersion() {
	fmt.Println("version:", version)
	fmt.Println("compiler:", runtime.Compiler, runtime.Version())
	fmt.Println("build:", buildDate)
	os.Exit(0)
}

func printError(msg string) {
	fmt.Println(msg)
	os.Exit(2)
}

type logger struct {
	*log.Logger
}

func (l *logger) Print(message string) {
	l.Logger.Output(3, message)
}
