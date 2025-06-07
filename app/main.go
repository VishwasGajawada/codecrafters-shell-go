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
	name         string
	args         []string
	outputStream *os.File
	errorStream  *os.File
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
func handleEcho(command *Command) {
	fmt.Fprintln(command.outputStream, strings.Join(command.args, " "))
}

// Handle the "type" command
func handleType(command *Command) {
	if isShellBuiltin(command.args[0]) {
		fmt.Fprintf(command.outputStream, "%s is a shell builtin\n", command.args[0])
		return
	}

	filePath, found := findExecutablePath(command.args[0])

	if found {
		fmt.Fprintf(command.outputStream, "%s is %s\n", command.args[0], filePath)
	} else {
		fmt.Fprintf(command.errorStream, "%s: not found\n", command.args[0])
	}
}

func handlePwd(command *Command) string {
	cwd, _ := os.Getwd()
	fmt.Fprintln(command.outputStream, cwd)
	return cwd
}

func handleCd(command *Command) {
	if command.args == nil || len(command.args) != 1 {
		fmt.Fprintln(command.errorStream, "cd: missing argument")
		return
	}
	absolutePath := getAbsolutePath(command.args[0])
	if isValidPath(absolutePath) {
		os.Chdir(absolutePath)
	} else {
		fmt.Fprintf(command.errorStream, "cd: %s: No such file or directory\n", absolutePath)
	}
}

func splitSingleWordByRedirects(input string) []string {
	var result []string
	var current string
	for i := 0; i < len(input); i++ {
		if input[i] == '>' || input[i] == '<' {
			if current != "" {
				result = append(result, current)
			}
			if current != "1" && current != "2" {
				result = append(result, "1") // Default output stream if not specified
			}
			current = ""
			// Handle multiple redirects
			if i+1 < len(input) && input[i] == input[i+1] {
				result = append(result, string(input[i])+string(input[i+1]))
				i++ // Skip the next character as it's part of the redirect
			} else {
				result = append(result, string(input[i]))
			}
		} else {
			current += string(input[i])
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func splitByQuotesAndRedirects(input string) []string {
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
			// Split current by >, >>, <, << and add the separated parts to the result
			// fmt.Println("current before redirect split:", current)
			currentAfterRedirects := splitSingleWordByRedirects(current)
			// fmt.Println("current after redirect split:", currentAfterRedirects)
			if len(currentAfterRedirects) > 0 {
				result = append(result, currentAfterRedirects...)
			}
			current = ""

		default:
			// Otherwise, just add the character to the current word
			current += string(curChar)
		}
	}

	if current != "" {
		currentAfterRedirects := splitSingleWordByRedirects(current)
		if len(currentAfterRedirects) > 0 {
			result = append(result, currentAfterRedirects...)
		}
	}
	return result
}

func getCommand(input string) *Command {
	// Split the input into words
	words := splitByQuotesAndRedirects(input)
	// fmt.Println("Words after splitting:", words)

	// Handle empty input
	if len(words) == 0 {
		return nil
	}

	// The first word is the command name, the rest are arguments
	commandName := words[0]
	var args []string
	var outputStream *os.File = nil
	var errorStream *os.File = nil
	var notIncludeInArgs = make([]bool, len(words))
	for i := 1; i < len(words); i++ {
		if words[i] == ">" || words[i] == ">>" {
			fileOpenBitMask := os.O_CREATE | os.O_WRONLY
			if words[i] == ">>" {
				fileOpenBitMask |= os.O_APPEND
			}

			if i+1 < len(words) {
				if words[i-1] == "1" {
					outputStream, _ = os.OpenFile(words[i+1], fileOpenBitMask, 0644)
				} else if words[i-1] == "2" {
					errorStream, _ = os.OpenFile(words[i+1], fileOpenBitMask, 0644)
				} else {
					fmt.Println("something went wrong")
				}
				notIncludeInArgs[i-1] = true // Mark the previous word as not an argument
				notIncludeInArgs[i] = true   // Mark the current word as not an argument
				notIncludeInArgs[i+1] = true // Mark the next word (filename) as not an argument

				i++ // Skip the next word as it is the filename
			} else {
				fmt.Fprintln(os.Stderr, "Error: '>>' requires a filename")
			}
		}
	}
	for i := 1; i < len(words); i++ {
		if !notIncludeInArgs[i] {
			args = append(args, words[i])
		}
	}

	if outputStream == nil {
		outputStream = os.Stdout // Default output stream
	}
	if errorStream == nil {
		errorStream = os.Stderr // Default error stream
	}

	return &Command{name: commandName, args: args, outputStream: outputStream, errorStream: errorStream}
}

// Process the command entered by the user
func processCommand(input string) bool {
	// Split the command into words
	command := getCommand(input)

	switch command.name {
	case "exit":
		return true // Exit the shell
	case "echo":
		handleEcho(command)
		return false // Continue the shell
	case "type":
		handleType(command)
		return false // Continue the shell
	case "pwd":
		// Handle "pwd" command
		handlePwd(command)
		return false
	case "cd":
		handleCd(command)
		return false

	default:
		_, found := findExecutablePath(command.name)
		if !found {
			fmt.Fprintf(command.outputStream, "%s: command not found\n", command.name)
			return false
		} else {
			var cmd = exec.Command(command.name, command.args...)
			cmd.Stdout = command.outputStream
			cmd.Stderr = command.errorStream
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
