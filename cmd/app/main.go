package main

import (
	"os"

	"gerrit-cli/cmd/app/command"
)

func main() {
	command.Run(os.Args)
}
