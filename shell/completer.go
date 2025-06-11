package shell

import (
	"fmt"

	"github.com/chzyer/readline"
	// Import the fsutil package for path management
)

// createReadlineCompleter generates and returns a readline.Completer instance.
// It takes the list of built-in commands to use for completion.
func CreateReadlineCompleter(builtIns []string) *readline.PrefixCompleter {
	var completers []readline.PrefixCompleterInterface
	// Dynamically create PrefixCompleter items for each built-in command
	for _, cmd := range builtIns {
		completers = append(completers, readline.PcItem(cmd))
	}
	// Create and return the PrefixCompleter
	return readline.NewPrefixCompleter(completers...)
}

type TabCompleter struct {
	builtIns         []string
	path_executables []string
}

func (t *TabCompleter) Do(line []rune, pos int) ([][]rune, int) {
	candidates := make([][]rune, 0)
	for _, cmd := range t.builtIns {
		if len(cmd) >= pos && string(line) == cmd[:pos] {
			candidates = append(candidates, []rune(cmd[pos:]+" ")) // Added space at the end
		}
	}
	// fmt.Println(len(t.path_executables))
	for _, cmd := range t.path_executables {
		if len(cmd) >= pos && string(line) == cmd[:pos] {
			candidates = append(candidates, []rune(cmd[pos:]+" ")) // Added space at the end
		}
	}

	if len(candidates) == 0 {
		// Ring the bell (alert) and return the current input
		fmt.Print("\x07")
		return nil, pos
	}
	return candidates, pos
}
