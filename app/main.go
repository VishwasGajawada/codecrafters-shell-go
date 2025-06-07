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
var paths = strings.Split(os.Getenv("PATH"), ":")

type Command struct {
	name string
	args []string
}

// Helpers

// Print the shell prompt
func printPrompt() {
	fmt.Fprint(os.Stdout, "$ ")
}

// Read user input from stdin
func readCommand() (string, error) {
	command, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(command), nil
}

func isValidPath(fullPath string) bool {
	// Check if the command exists in the given path
	if _, err := os.Stat(fullPath); err == nil {
		return true
	} else {
		return false
	}
}

func getAbsolutePath(path string) string {
	if path == "" {
		return path
	} else if path == "~" {
		homeDir := os.Getenv("HOME")
		return homeDir
	} else if path[0] == '/' {
		return path
	} else {
		directory_changes := strings.Split(path, "/")
		cwd, _ := os.Getwd()
		curDirectories := strings.Split(cwd, "/")

		for _, change := range directory_changes {
			if change == ".." {
				// Pop the last directory if not at the root
				if len(curDirectories) > 0 {
					curDirectories = curDirectories[:len(curDirectories)-1]
				}
			} else if change != "." && change != "" {
				// Push the directory to the current path
				curDirectories = append(curDirectories, change)
			}
		}
		// Join the directories back into a single path
		absolutePath := strings.Join(curDirectories, "/")
		if !strings.HasPrefix(absolutePath, "/") {
			absolutePath = "/" + absolutePath
		}
		return absolutePath
	}
}

func findExecutablePath(command string) (string, bool) {
	// Check if the command exists in the given paths
	for _, dir := range paths {
		fullPath := fmt.Sprintf("%s/%s", dir, command)
		if isValidPath(fullPath) {
			return fullPath, true
		}
	}
	return "", false
}

func isShellBuiltin(command string) bool {
	// Check if the command is a shell builtin
	if slices.Contains(builtIns, command) {
		return true
	}
	return false
}

// Handle the "echo" command
func handleEcho(words []string) {
	fmt.Fprintln(os.Stdout, strings.Join(words, " "))
}

// Handle the "type" command
func handleType(word string) {
	if isShellBuiltin(word) {
		fmt.Fprintf(os.Stdout, "%s is a shell builtin\n", word)
		return
	}

	filePath, found := findExecutablePath(word)

	if found {
		fmt.Fprintf(os.Stdout, "%s is %s\n", word, filePath)
	} else {
		fmt.Fprintf(os.Stdout, "%s: not found\n", word)
	}
}

func handlePwd() string {
	cwd, _ := os.Getwd()
	fmt.Fprintln(os.Stdout, cwd)
	return cwd
}

func handleCd(path string) {
	absolutePath := getAbsolutePath(path)
	if isValidPath(absolutePath) {
		os.Chdir(absolutePath)
	} else {
		fmt.Fprintf(os.Stderr, "cd: %s: No such file or directory\n", absolutePath)
	}
}

func splitByQuotes(input string) []string {
	var result []string
	var current string
	inSingleQuotes, inDoubleQuotes, escaped := false, false, false
	for i := range len(input) {
		var curChar = input[i]

		switch {
		case escaped:
			// If the current character is escaped, just add it to the current string
			current += string(input[i])
			escaped = false

		case curChar == '\'' && !inDoubleQuotes:
			inSingleQuotes = !inSingleQuotes

		case curChar == '"' && !inSingleQuotes:
			inDoubleQuotes = !inDoubleQuotes

		case curChar == '\\' && !inSingleQuotes && !inDoubleQuotes:
			escaped = true

		case curChar == '\\' && inDoubleQuotes:
			// escape next character if it is $, " or \
			if (i+1) < len(input) && (input[i+1] == '$' || input[i+1] == '"' || input[i+1] == '\\') {
				escaped = true
			} else {
				current += string(curChar) // Just add the backslash if not escaping
			}

		case curChar == ' ' && !inSingleQuotes && !inDoubleQuotes:
			// If we encounter a space and not in quotes, finalize the current word
			if current != "" {
				result = append(result, current)
				current = ""
			}

		default:
			// Otherwise, just add the character to the current word
			current += string(curChar)
		}
	}

	if current != "" {
		result = append(result, current)
	}
	return result
}

func getCommand(input string) *Command {
	// Split the input into words
	words := splitByQuotes(input)

	// Handle empty input
	if len(words) == 0 {
		return nil
	}

	// The first word is the command name, the rest are arguments
	commandName := words[0]
	args := words[1:]

	return &Command{name: commandName, args: args}
}

// Process the command entered by the user
func processCommand(input string) bool {
	// Split the command into words
	command := getCommand(input)

	switch command.name {
	case "exit":
		return true // Exit the shell
	case "echo":
		handleEcho(command.args)
		return false // Continue the shell
	case "type":
		handleType(command.args[0])
		return false // Continue the shell
	case "pwd":
		// Handle "pwd" command
		handlePwd()
		return false
	case "cd":
		if command.args == nil || len(command.args) != 1 {
			fmt.Fprintln(os.Stderr, "cd: missing argument")
			return false
		}
		handleCd(command.args[0]) //only absolute paths are supported
		return false

	default:
		_, found := findExecutablePath(command.name)
		if !found {
			fmt.Fprintf(os.Stderr, "%s: command not found\n", command.name)
			return false
		} else {
			var cmd = exec.Command(command.name, command.args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
		return false
	}
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
