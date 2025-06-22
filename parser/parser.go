package parser

import (
	"fmt"
	"os"
	"strings"
	"unicode"

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

func splitByPipes(input string) []string {
	var (
		result         []string
		current        string
		inSingleQuotes bool
		inDoubleQuotes bool
		escaped        bool
	)
	for i := range len(input) {
		c := input[i]
		switch {
		case escaped:
			current += string(c)
			escaped = false
		case c == '\\' && !inSingleQuotes:
			escaped = true
			current += string(c) // keep escape character in first token
		case c == '\'' && !inDoubleQuotes:
			inSingleQuotes = !inSingleQuotes
			current += string(c)
		case c == '"' && !inSingleQuotes:
			inDoubleQuotes = !inDoubleQuotes
			current += string(c)
		case c == '|' && !inSingleQuotes && !inDoubleQuotes:
			// unquoted pipe – split here
			if current != "" {
				result = append(result, current)
				current = ""
			}
		default:
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// The first token (command) retains all quotes/escapes; subsequent tokens are stripped and unescaped before redirect splitting.
func splitByQuotesAndRedirects(input string) []string {
	input = strings.TrimLeftFunc(input, unicode.IsSpace)
	var (
		result         []string
		current        string
		inSingleQuotes bool
		inDoubleQuotes bool
		escaped        bool
		firstDone      bool
	)
	runes := []rune(input)
	for i, r := range runes {

		// fmt.Printf("Processing rune: %c (index %d)\n", r, i) // Debugging output
		switch {
		case escaped:
			// fmt.Printf("Escaped character: ~%c~\n", r) // Debugging output
			current += string(r)
			escaped = false
		case r == '\\' && !inSingleQuotes:
			if !inDoubleQuotes {
				escaped = true
				if !firstDone {
					current += string(r) // keep escape character in first token
				}
			} else {
				if i+1 < len(input) && (input[i+1] == ' ' ||
					input[i+1] == '\t' ||
					input[i+1] == '|' ||
					input[i+1] == '<' ||
					input[i+1] == '>' ||
					input[i+1] == '\\' ||
					(!inDoubleQuotes && input[i+1] == '\'') ||
					input[i+1] == '"' ||
					input[i+1] == '\n') {
					escaped = true
					if !firstDone {
						current += string(r)
					}
				} else {
					current += string(r)
				}
			}
		case r == '\'' && !inDoubleQuotes:
			inSingleQuotes = !inSingleQuotes
			if !firstDone {
				current += string(r) // keep escape character in first token
			}
		case r == '"' && !inSingleQuotes:
			inDoubleQuotes = !inDoubleQuotes
			if !firstDone {
				current += string(r) // keep escape character in first token
			}
		case unicode.IsSpace(r) && !inSingleQuotes && !inDoubleQuotes:
			if current != "" {
				if !firstDone {
					// first token: keep quotes
					result = append(result, current)
					firstDone = true
				} else {
					// strip surrounding quotes for subsequent tokens
					stripped := StripSurroundingQuotes(current)
					result = append(result, splitSingleWordByRedirects(stripped)...)
				}
			}
			current = ""
		default:
			current += string(r)
		}
		// flush at end
		if i == len(runes)-1 && current != "" {
			if !firstDone {
				result = append(result, current)
			} else {
				stripped := StripSurroundingQuotes(current)
				result = append(result, splitSingleWordByRedirects(stripped)...)
			}
		}

		// fmt.Printf("Current token: ~%s~\n", current) // Debugging output
	}
	return result
}

// stripSurroundingQuotes removes matching single or double quotes around a word.
func StripSurroundingQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func ParseCommand(input string, curInputStream *os.File, curOutputStream *os.File) *types.Command {
	// Split the input into words
	words := splitByQuotesAndRedirects(input)
	// fmt.Printf("RAW  (%d bytes): %s\n", len(input), input)
	// fmt.Printf("GO-LIT: %#v\n", words)

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
		outputStream = curOutputStream // Default output stream
	}
	if errorStream == nil {
		errorStream = os.Stderr // Default error stream
	}

	return &types.Command{Name: commandName, Args: args, InputStream: curInputStream, OutputStream: outputStream, ErrorStream: errorStream}
}

// Splits input by Pipe characters
func GetCommands(input string) []string {
	commandStrings := splitByPipes(input)
	return commandStrings
}
