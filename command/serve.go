/* For license and copyright information please see the LEGAL file in the code repository */

package cmd

import (
	"fmt"
	"os"

	"github.com/GeniusesGroup/libgo/protocol"
)

func ServeCLA(c protocol.Command, arguments []string) (err protocol.Error) {
	var serviceName string
	if len(arguments) > 0 {
		serviceName = arguments[0]
	} else {
		serviceName = "help"
	}

	// Also check for finding help command to check any custom help command
	var command protocol.Command = c.SubCommand(serviceName)
	if command == nil {
		// We don't find any related command even custom help, so print auto generated help.
		if serviceName == "help" || serviceName == "-h" || serviceName == "--help" {
			// helpMessage()
			// Accept 'go mod help' and 'go mod help foo' for 'go help mod' and 'go help mod foo'.
			// help.Help(os.Stdout, append(strings.Split(cfg.CmdName, " "), args[1:]...))
		} else {
			fmt.Fprintf(os.Stderr, "unknown command\nRun '%s help' for usage.\n", CommandPath(c))
			err = &ErrServiceNotFound
		}
		return
	} else if command.Name() != serviceName {
		fmt.Fprintf(os.Stderr, "Do you mean '%s %s'?\n", CommandPath(c), command.Name())
		// TODO:::
		return
	}

	err = command.ServeCLA(arguments[1:])
	return
}

// Root finds root command. or return nil if it is the root
func Root(c protocol.Command) (root protocol.Command) {
	for {
		if root.Parent() != nil {
			root = root.Parent()
		} else {
			break
		}
	}
	return
}

// CommandPath returns the full path to this command exclude itself.
func CommandPath(command protocol.Command) (fullName string) {
	for {
		fullName = command.Name() + " " + fullName
		command = command.Parent()
		if command == nil {
			break
		}
	}
	// remove trailing space
	fullName = fullName[:len(fullName)-1]
	return
}
