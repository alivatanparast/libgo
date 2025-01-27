/* For license and copyright information please see the LEGAL file in the code repository */

package main

import (
	"os"

	"github.com/GeniusesGroup/libgo/modules"
)

func init() {
	modules.RootCommand.Init()
}

func main() {
	// remove app binary path from os args
	var args = os.Args[1:]
	var err = modules.RootCommand.ServeCLA(args)
	if err != nil {
		os.Exit(1)
	}
}
