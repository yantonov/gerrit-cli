package command

import (
	"fmt"
	"os"
)

type Command struct {
	usage          string
	requiresClient bool
	run            func(*GerritClient, []string) error
}

var commandNames = []string{
	"get-change",
	"get-files",
	"get-commit",
	"get-diff",
	"get-messages",
	"get-patch",
	"get-moab-numbers",
	"post-comment",
	"resolve-change-number",
	"resolve-change-id",
}

var commands = map[string]Command{
	"get-change":            getChangeCommand,
	"get-files":             getFilesCommand,
	"get-commit":            getCommitCommand,
	"get-diff":              getDiffCommand,
	"get-messages":          getMessagesCommand,
	"get-patch":             getPatchCommand,
	"get-moab-numbers":      getMoabNumbersCommand,
	"post-comment":          postCommentCommand,
	"resolve-change-number": resolveChangeNumberCommand,
	"resolve-change-id":     resolveChangeIDCommand,
}

func Run(args []string) {
	programName := "gerrit-cli"
	if len(args) > 0 {
		programName = args[0]
	}

	if len(args) < 2 {
		printUsage(programName)
		os.Exit(1)
	}

	commandName := args[1]
	commandArgs := args[2:]

	if commandName == "resolve-change-number" {
		runCommand(commands[commandName], nil, commandArgs)
		return
	}

	client, err := NewGerritClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	command, ok := commands[commandName]
	if !ok {
		fmt.Fprintf(os.Stderr, "ERROR: Unknown command '%s'\n", commandName)
		os.Exit(1)
	}

	runCommand(command, client, commandArgs)
}

func runCommand(command Command, client *GerritClient, args []string) {
	if err := command.run(client, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printUsage(programName string) {
	fmt.Fprintf(os.Stderr, "Usage: %s <command> [args...]\n", programName)
	fmt.Fprintf(os.Stderr, "Commands:\n")
	for _, name := range commandNames {
		fmt.Fprintf(os.Stderr, "  %s\n", commands[name].usage)
	}
}
