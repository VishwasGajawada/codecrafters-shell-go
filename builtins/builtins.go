// CODECRAFTERS-SHELL-GO/builtin/builtin.go
package builtin

import (
	"fmt"
	"os"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/fsutil"
	"github.com/codecrafters-io/shell-starter-go/types" // Import the new types package
)

// HandleEcho handles the "echo" command.
func HandleEcho(command *types.Command) { // Parameter type changed
	fmt.Fprintln(command.OutputStream, strings.Join(command.Args, " "))
}

// HandleType handles the "type" command.
func HandleType(command *types.Command, pathFinder *fsutil.Finder, builtins []string) { // Parameter type changed
	// ... rest of function using command.Args, command.OutputStream, command.ErrorStream
	if len(command.Args) == 0 {
		fmt.Fprintln(command.ErrorStream, "type: missing argument")
		return
	}

	cmdName := command.Args[0]
	// commonBuiltins := []string{"echo", "type", "exit", "pwd", "cd"}
	for _, b := range builtins {
		if cmdName == b {
			fmt.Fprintf(command.OutputStream, "%s is a shell builtin\n", cmdName)
			return
		}
	}

	filePath, found := pathFinder.FindExecutablePath(cmdName)
	if found {
		fmt.Fprintf(command.OutputStream, "%s is %s\n", cmdName, filePath)
	} else {
		fmt.Fprintf(command.ErrorStream, "%s: not found\n", cmdName)
	}
}

// HandlePwd handles the "pwd" command.
func HandlePwd(command *types.Command) { // Parameter type changed
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(command.ErrorStream, "pwd: error getting current directory: %v\n", err)
		return
	}
	fmt.Fprintln(command.OutputStream, cwd)
}

// HandleCd handles the "cd" command.
func HandleCd(command *types.Command, pathFinder *fsutil.Finder) { // Parameter type changed
	if len(command.Args) == 0 || len(command.Args) > 1 {
		fmt.Fprintln(command.ErrorStream, "cd: missing or too many arguments")
		return
	}

	targetPath := command.Args[0]
	absolutePath := pathFinder.GetAbsolutePath(targetPath)

	if pathFinder.IsValidPath(absolutePath) {
		err := os.Chdir(absolutePath)
		if err != nil {
			fmt.Fprintf(command.ErrorStream, "cd: %s: %v\n", targetPath, err)
		}
	} else {
		fmt.Fprintf(command.ErrorStream, "cd: %s: No such file or directory\n", targetPath)
	}
}
