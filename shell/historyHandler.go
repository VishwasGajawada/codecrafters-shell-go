package shell

import (
	"bufio"
	"fmt"
	"os"

	"github.com/codecrafters-io/shell-starter-go/fsutil"
	"github.com/codecrafters-io/shell-starter-go/types"
)

func GetHistoryFromFile(historyFilePath string) ([]string, error) {
	if !fsutil.IsValidPath(historyFilePath) {
		return nil, fmt.Errorf("invalid file path: %s", historyFilePath)
	}

	file, err := os.Open(historyFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %v", historyFilePath, err)
	}
	defer file.Close()

	var history []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			history = append(history, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", historyFilePath, err)
	}

	return history, nil
}

func (s *Shell) LoadHistoryFromFile(command *types.Command, historyFilePath string) {
	historyLoaded, err := GetHistoryFromFile(historyFilePath)
	if err != nil {
		fmt.Fprintf(command.ErrorStream, "history: %v\n", err)
		return
	}

	s.CommandsHistory = append(s.CommandsHistory, historyLoaded...)
}

func (s *Shell) WriteHistoryToFile(command *types.Command, filePath string, append bool) {
	fileOpenBitMask := os.O_CREATE | os.O_WRONLY
	if append {
		fileOpenBitMask |= os.O_APPEND
	}

	file, err := os.OpenFile(filePath, fileOpenBitMask, 0644)
	if err != nil {
		fmt.Fprintf(command.ErrorStream, "history: error opening file %s: %v\n", filePath, err)
		return
	}
	defer file.Close()

	for _, cmd := range s.CommandsHistory {
		if cmd != "" {
			_, err := fmt.Fprintln(file, cmd)
			if err != nil {
				fmt.Fprintf(command.ErrorStream, "history: error writing to file %s: %v\n", filePath, err)
				return
			}
		}
	}
}

// HISTFILE env variable
func GetHistoryFromEnv() []string {
	historyFilePath := os.Getenv("HISTFILE")
	// fmt.Printf("HISTFILE: %s\n", historyFilePath) // Debugging output
	if historyFilePath == "" {
		return make([]string, 0)
	}

	if !fsutil.IsValidPath(historyFilePath) {
		return make([]string, 0)
	}

	history, err := GetHistoryFromFile(historyFilePath)
	// fmt.Printf("Found number of history entries: %d\n", len(history)) // Debugging output
	if err != nil {
		fmt.Fprintf(os.Stderr, "history: %v\n", err)
		return make([]string, 0)
	}
	return history
}

func (s *Shell) WriteHistoryToEnv() {
	historyFilePath := os.Getenv("HISTFILE")
	if historyFilePath == "" {
		return
	}

	if !fsutil.IsValidPath(historyFilePath) {
		return
	}

	s.WriteHistoryToFile(&types.Command{OutputStream: os.Stdout, ErrorStream: os.Stderr}, historyFilePath, true)
}
