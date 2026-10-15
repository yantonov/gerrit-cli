package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Command struct {
	usage          string
	requiresClient bool
	run            func(*GerritClient, []string) error
}

var commandNames = []string{
	"keychain",
	"detail",
	"files",
	"commit",
	"diff",
	"messages",
	"patch",
	"moab-numbers",
	"publish-version",
	"is-verified",
	"query",
	"review",
	"resolve-change-number",
	"resolve-change-id",
	"shell",
}

var commands = map[string]Command{
	"keychain":              keychainCommand,
	"detail":                detailCommand,
	"files":                 filesCommand,
	"commit":                commitCommand,
	"diff":                  diffCommand,
	"messages":              messagesCommand,
	"patch":                 patchCommand,
	"moab-numbers":      getMoabNumbersCommand,
	"publish-version":   getPublishVersionCommand,
	"is-verified":           isVerifiedCommand,
	"query":                 queryCommand,
	"review":                reviewCommand,
	"resolve-change-number": resolveChangeNumberCommand,
	"resolve-change-id":     resolveChangeIDCommand,
	"shell":                 shellCommand,
}

const defaultProgramName = "gerrit-cli"

// programName is what the usage lines and the shell completion script tell
// the user to type: the binary as it was invoked, without the directory it
// was found in. Windows invokes it with the .exe extension, which is not
// part of the name to type.
func programName(args []string) string {
	if len(args) == 0 {
		return defaultProgramName
	}

	name := filepath.Base(args[0])
	if extension := filepath.Ext(name); strings.EqualFold(extension, ".exe") {
		name = strings.TrimSuffix(name, extension)
	}

	// filepath.Base answers "." for an empty path and the separator for a bare
	// root, neither of which is a name anyone can type.
	if name == "" || name == "." || name == string(filepath.Separator) {
		return defaultProgramName
	}

	return name
}

func Run(args []string) {
	program := programName(args)

	if len(args) < 2 {
		printUsage(program)
		os.Exit(1)
	}

	commandName := args[1]
	commandArgs := args[2:]

	if commandName == "--help" || commandName == "-h" || commandName == "help" {
		printUsage(program)
		return
	}

	if commandName == "--version" || commandName == "version" {
		printVersion()
		return
	}

	command, ok := commands[commandName]
	if !ok {
		fmt.Fprintf(os.Stderr, "ERROR: Unknown command '%s'\n", commandName)
		os.Exit(1)
	}

	if !command.requiresClient {
		runCommand(command, nil, commandArgs)
		return
	}

	client, err := NewGerritClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
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
