package shell

import (
	"github.com/chzyer/readline"
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
