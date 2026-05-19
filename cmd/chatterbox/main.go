package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: chatterbox <run|scenario|version> [flags]\n\nRun chatterbox run -h for help")
	}
	switch args[0] {
	case "scenario":
		if len(args) < 2 {
			return fmt.Errorf("usage: chatterbox scenario run -f <file.yaml>")
		}
		if args[1] != "run" {
			return fmt.Errorf("unknown scenario subcommand %q (use: scenario run)", args[1])
		}
		cfg, err := parseScenarioRunFlags(args[2:])
		if err != nil {
			if err == errHelp {
				return nil
			}
			return err
		}
		return executeScenarioRun(cfg)
	case "run":
		cfg, err := parseRunFlags(args[1:])
		if err != nil {
			if err == errHelp {
				return nil
			}
			return err
		}
		return executeRun(cfg)
	case "version", "-v", "--version":
		return runVersion()
	case "-h", "--help", "help":
		printRootUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q (use run or version)", args[0])
	}
}

var errHelp = fmt.Errorf("help")

func printRootUsage() {
	fmt.Fprintf(os.Stderr, "Chatterbox — synthetic fuzzy logs for pipeline testing\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n  chatterbox run [flags]              generate logs\n  chatterbox scenario run -f FILE     run failure scenario\n  chatterbox version                  print version\n")
}
