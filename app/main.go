package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// Print the shell prompt
func printPrompt() {
	fmt.Fprint(os.Stdout, "$ ")
}

var builtIns = []string{"echo", "type", "exit"}

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

func commandFoundInPath(command string, path string) bool {
	// Check if the command exists in the given path
	fullPath := fmt.Sprintf("%s/%s", path, command)
	if _, err := os.Stat(fullPath); err == nil {
		return true
	} else {
		return false
	}
}

// Handle the "type" command
func handleType(words []string) {
	if len(words) == 2 {
		var path = os.Getenv("PATH")
		var validCommand bool = false
		var dirFound string = ""

		if slices.Contains(builtIns, words[1]) {
			fmt.Fprintf(os.Stdout, "%s is a shell builtin\n", words[1])
			return
		}

		for _, dir := range strings.Split(path, ":") {
			if commandFoundInPath(words[1], dir) {
				validCommand = true
				dirFound = dir
				break
			}
		}

		if validCommand {
			fmt.Fprintf(os.Stdout, "%s is %s/%s\n", words[1], dirFound, words[1])
		} else {
			fmt.Fprintf(os.Stdout, "%s: not found\n", words[1])
		}
	}
}

// Process the command entered by the user
func processCommand(command string) bool {
	// Split the command into words
	words := strings.Fields(command)

	// Handle empty input
	if len(words) == 0 {
		return false
	}

	// Handle "exit" command
	if words[0] == "exit" {
		return true
	}

	// Handle "echo" command
	if words[0] == "echo" {
		handleEcho(words)
		return false
	}

	// Handle "type" command
	if words[0] == "type" {
		handleType(words)
		return false
	}

	var output []byte
	// Handle unknown commands
	if len(words) > 1 {
		output, _ = exec.Command(words[0], strings.Join(words[1:], " ")).Output()
	} else {
		output, _ = exec.Command(words[0]).Output()
	}

	fmt.Println(strings.TrimSpace(string(output)))
	return false
}

func main() {

	for {
		printPrompt()

		// Read the command
		command, err := readCommand()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading command: %v\n", err)
			break
		}

		// Process the command
		if processCommand(command) {
			break
		}
	}
}
