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

func findExecutablePath(command string, paths []string) (string, bool) {
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
	if len(words) > 1 {
		fmt.Fprintln(os.Stdout, strings.Join(words[1:], " "))
	} else {
		fmt.Fprintln(os.Stdout, "")
	}
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
	inSingleQuotes, inDoubleQuotes := false, false
	for i := 0; i < len(input); i++ {
		var curChar = input[i]
		if curChar == '\'' && !inDoubleQuotes {
			inSingleQuotes = !inSingleQuotes
		} else if curChar == '"' && !inSingleQuotes {
			inDoubleQuotes = !inDoubleQuotes
		} else if curChar == '\\' && (i+1) < len(input) {
			nextChar := input[i+1]
			if nextChar == '$' || nextChar == '"' || nextChar == '\\' {
				current += string(nextChar)
				i++ // Skip the next character since it's escaped
			} else {
				current += string(curChar) // Just add the backslash
			}
		} else if curChar == ' ' && !inSingleQuotes && !inDoubleQuotes {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(curChar)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// Process the command entered by the user
func processCommand(input string, paths []string) bool {
	// Split the command into words
	words := splitByQuotes(input)

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
		if args == nil || len(args) != 1 {
			fmt.Fprintln(os.Stderr, "cd: missing argument")
			return false
		}
		handleCd(args[0]) //only absolute paths are supported
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
}

func main() {
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
