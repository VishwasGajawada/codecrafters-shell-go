// my-go-shell/shell/shell.go
package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/chzyer/readline"
	builtin "github.com/codecrafters-io/shell-starter-go/builtins" // Import builtin package
	"github.com/codecrafters-io/shell-starter-go/fsutil"           // Import fsutil package
	"github.com/codecrafters-io/shell-starter-go/parser"           // Import parser package
	"github.com/codecrafters-io/shell-starter-go/trie"             // Import trie package for autocompletion
	"github.com/codecrafters-io/shell-starter-go/types"            // Import types package
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

	allCommands := make([]string, 0)
	allCommands = append(allCommands, builtIns...)
	allCommands = append(allCommands, pathFinder.GetExecutables()...) // Get executables from PATH

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "$ ",
		InterruptPrompt: "^C",   // Text to show on Ctrl+C
		EOFPrompt:       "exit", // Text to show on Ctrl+D
		AutoComplete: &TabCompleter{
			trie:                           trie.NewTrieNode(allCommands), // Initialize trie with all commands
			tabPressedAfterMultipleResults: false,
		},
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
		exitShell := s.processInput(commandInput)
		if exitShell {
			break
		}
	}
}

// printPrompt prints the shell prompt to stdout.
func (s *Shell) printPrompt() {
	fmt.Fprint(os.Stdout, "$ ")
}

func (s *Shell) processInput(input string) bool {
	commandStrings := parser.GetCommands(input)
	// fmt.Fprintf(os.Stdout, "commandStrings: %v\n", commandStrings) // Debugging output

	inputStreams := make([]*os.File, len(commandStrings))
	outputStreams := make([]*os.File, len(commandStrings))

	// Set default input and output streams for each command
	for i := range commandStrings {
		inputStreams[i] = os.Stdin
		outputStreams[i] = os.Stdout
	}

	// Connect output streams of previous commands to input streams of next commands
	for i := 0; i < len(commandStrings)-1; i++ {
		pipeReader, pipeWriter, err := os.Pipe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating pipe: %v\n", err)
			return false // Continue the shell, but log the error
		}
		inputStreams[i+1] = pipeReader // Set the next command's input to the pipe reader
		outputStreams[i] = pipeWriter  // Set the current command's output to the pipe writer
	}

	commands := make([]*types.Command, len(commandStrings))
	for i, cmdStr := range commandStrings {
		commands[i] = parser.ParseCommand(cmdStr, inputStreams[i], outputStreams[i])
	}

	var wgExecute sync.WaitGroup
	exitCodes := make([]bool, len(commands))
	for idx, cmd := range commands {
		wgExecute.Add(1)
		go func(cmd *types.Command) {
			defer wgExecute.Done()
			exitCodes[idx] = s.processCommand(cmd)
		}(cmd)
	}
	wgExecute.Wait() // Wait for all commands to finish executing

	return exitCodes[len(exitCodes)-1] // Return the exit code of the last command
}

// processCommand parses and executes a command. Returns true if the shell should exit.
func (s *Shell) processCommand(cmd *types.Command) bool {
	if cmd == nil { // Handle cases where parser returns nil (e.g., only redirects or empty)
		return false
	}

	defer func() {
		// Close redirected files if they were opened
		if cmd.InputStream != os.Stdin {
			cmd.InputStream.Close()
		}
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
	strippedCommand := parser.StripSurroundingQuotes(cmd.Name) // Strip quotes for command name
	path, found := s.pathFinder.FindExecutablePath(strippedCommand)
	if !found {
		fmt.Fprintf(cmd.ErrorStream, "%s: command not found\n", cmd.Name)
		return
	}

	commandToRun := cmd.Name
	if strippedCommand != cmd.Name {
		commandToRun = path // Use the full path if command was quoted
	}
	execCmd := exec.Command(commandToRun, cmd.Args...)
	execCmd.Stdout = cmd.OutputStream
	execCmd.Stderr = cmd.ErrorStream
	execCmd.Stdin = cmd.InputStream
	execCmd.Run()
}
