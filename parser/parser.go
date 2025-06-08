package parser

import (
	"fmt"
	"os"

	"github.com/codecrafters-io/shell-starter-go/types" // Import the shell package to use its Command struct
)

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

func GetCommand(input string) *types.Command {
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

	return &types.Command{Name: commandName, Args: args, OutputStream: outputStream, ErrorStream: errorStream}
}
