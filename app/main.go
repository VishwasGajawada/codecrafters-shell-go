package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// Globals
var builtIns = []string{"echo", "type", "exit", "pwd", "cd"}
var currentWorkingDir string

// Print the shell prompt
func printPrompt() {
	fmt.Fprint(os.Stdout, "$ ")
}

func isShellBuiltin(command string) bool {
	// Check if the command is a shell builtin
	if slices.Contains(builtIns, command) {
		return true
	}
	return false
}

// Read user input from stdin
func readCommand() (string, error) {
	command, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(command), nil
}

// Handle the "echo" command
func handleEcho(words []string) {
	if len(words) > 1 {
		fmt.Fprintln(os.Stdout, strings.Join(words[1:], " "))
	} else {
		fmt.Fprintln(os.Stdout, "")
	}
}

func commandFoundInPath(fullPath string, path string) bool {
	// Check if the command exists in the given path
	if _, err := os.Stat(fullPath); err == nil {
		return true
	} else {
		return false
	}
}

func findExecutablePath(command string, paths []string) (string, bool) {
	// Check if the command exists in the given paths
	for _, dir := range paths {
		fullPath := fmt.Sprintf("%s/%s", dir, command)
		if commandFoundInPath(fullPath, dir) {
			return fullPath, true
		}
	}
	return "", false
}

// Handle the "type" command
func handleType(words []string, paths []string) {
	if len(words) == 2 {
		if isShellBuiltin(words[1]) {
			fmt.Fprintf(os.Stdout, "%s is a shell builtin\n", words[1])
			return
		}

		filePath, found := findExecutablePath(words[1], paths)

		if found {
			fmt.Fprintf(os.Stdout, "%s is %s\n", words[1], filePath)
		} else {
			fmt.Fprintf(os.Stdout, "%s: not found\n", words[1])
		}
	}
}
func handlePwd() {
	fmt.Fprintln(os.Stdout, currentWorkingDir)
}

// Process the command entered by the user
func processCommand(input string, paths []string) bool {
	// Split the command into words
	words := strings.Fields(input)

	// Handle empty input
	if len(words) == 0 {
		return false
	}

	command := words[0]
	args := words[1:]

	switch command {
	case "exit":
		return true // Exit the shell
	case "echo":
		handleEcho(words)
		return false // Continue the shell
	case "type":
		handleType(words, paths)
		return false // Continue the shell
	case "pwd":
		// Handle "pwd" command
		handlePwd()
		return false
	case "cd":
		currentWorkingDir = args[0] //only absolute paths are supported
		return false

	default:
		_, found := findExecutablePath(words[0], paths)
		if !found {
			fmt.Fprintf(os.Stderr, "%s: command not found\n", words[0])
			return false
		} else {
			var cmd = exec.Command(words[0], args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
		return false
	}
	return false
}

func main() {
	currentWorkingDir, _ = os.Getwd()
	var path = os.Getenv("PATH")
	var paths []string = strings.Split(path, ":")
	for {
		printPrompt()

		// Read the command
		command, err := readCommand()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading command: %v\n", err)
			break
		}

		// Process the command
		if processCommand(command, paths) {
			break
		}
	}
}
