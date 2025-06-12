// my-go-shell/shell/shell.go
package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/chzyer/readline"
	builtin "github.com/codecrafters-io/shell-starter-go/builtins" // Import builtin package
	"github.com/codecrafters-io/shell-starter-go/fsutil"           // Import fsutil package
	"github.com/codecrafters-io/shell-starter-go/parser"           // Import parser package
	"github.com/codecrafters-io/shell-starter-go/types"            // Import parser package
)

// Shell encapsulates the state and behavior of the shell.
type Shell struct {
	builtIns   []string
	pathFinder *fsutil.Finder // Use a struct for path management
	rl         *readline.Instance
}

// NewShell creates and initializes a new Shell instance.
func NewShell() *Shell {
	builtIns := []string{"echo", "type", "exit", "pwd", "cd"}
	pathFinder := fsutil.NewFinder(strings.Split(os.Getenv("PATH"), ":")) // Initialize path finder

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "$ ",
		InterruptPrompt: "^C",   // Text to show on Ctrl+C
		EOFPrompt:       "exit", // Text to show on Ctrl+D
		AutoComplete: &TabCompleter{
			builtIns:         builtIns,
			// path_executables: make([]string, 0), // Initialize with an empty slice
			path_executables:               pathFinder.GetExecutables(),
			tabPressedAfterMultipleResults: false}, // <--- Assign our custom completer
		// AutoComplete:    CreateReadlineCompleter(builtIns), // <--- Assign our custom completer
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing readline: %v\n", err)
		os.Exit(1) // Cannot run interactive shell without readline
	}
	return &Shell{
		builtIns:   builtIns,
		pathFinder: pathFinder,
		rl:         rl,
	}
}

// Run starts the shell's main loop.
func (s *Shell) Run() {
	defer s.rl.Close() // Ensure readline is closed when done

	for {
		s.printPrompt()

		commandInput, err := s.rl.Readline()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading command: %v\n", err)
			break
		}

		if commandInput == "" { // Handle empty input gracefully
			continue
		}

		// Process the command
		exitShell := s.processCommand(commandInput)
		if exitShell {
			break
		}
	}
}

// printPrompt prints the shell prompt to stdout.
func (s *Shell) printPrompt() {
	fmt.Fprint(os.Stdout, "$ ")
}

// processCommand parses and executes a command. Returns true if the shell should exit.
func (s *Shell) processCommand(input string) bool {
	// Parse the command input using the parser package
	cmd := parser.GetCommand(input)
	if cmd == nil { // Handle cases where parser returns nil (e.g., only redirects or empty)
		return false
	}

	defer func() {
		// Close redirected files if they were opened
		if cmd.OutputStream != os.Stdout && cmd.OutputStream != os.Stderr {
			cmd.OutputStream.Close()
		}
		if cmd.ErrorStream != os.Stdout && cmd.ErrorStream != os.Stderr {
			cmd.ErrorStream.Close()
		}
	}()

	switch cmd.Name {
	case "exit":
		return true // Exit the shell
	case "echo":
		builtin.HandleEcho(cmd)
	case "type":
		builtin.HandleType(cmd, s.pathFinder, s.builtIns) // Pass the pathFinder instance
	case "pwd":
		builtin.HandlePwd(cmd)
	case "cd":
		builtin.HandleCd(cmd, s.pathFinder) // Pass the pathFinder instance
	default:
		// Attempt to execute as an external command
		s.executeExternalCommand(cmd)
	}
	return false // Continue the shell
}

// executeExternalCommand finds and runs an external command.
func (s *Shell) executeExternalCommand(cmd *types.Command) {
	_, found := s.pathFinder.FindExecutablePath(cmd.Name)
	if !found {
		fmt.Fprintf(cmd.ErrorStream, "%s: command not found\n", cmd.Name)
		return
	}

	execCmd := exec.Command(cmd.Name, cmd.Args...)
	execCmd.Stdout = cmd.OutputStream
	execCmd.Stderr = cmd.ErrorStream
	execCmd.Stdin = os.Stdin // External commands should inherit stdin by default
	execCmd.Run()
}
