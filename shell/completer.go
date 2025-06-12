package shell

import (
	"fmt"
	"sort"

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
	builtIns                       []string
	path_executables               []string
	tabPressedAfterMultipleResults bool
	lastEnteredLine                string
}

func ringBell() {
	// Ring the bell (alert) to indicate no completion found
	fmt.Print("\x07")
}

func getUniqueStrings(strs []string) []string {
	// Create a map to track unique strings
	uniqueMap := make(map[string]struct{})
	for _, str := range strs {
		uniqueMap[str] = struct{}{}
	}

	// Convert the map keys back to a slice
	uniqueStrs := make([]string, 0, len(uniqueMap))
	for str := range uniqueMap {
		uniqueStrs = append(uniqueStrs, str)
	}
	return uniqueStrs
}

func (t *TabCompleter) Do(line []rune, pos int) ([][]rune, int) {
	candidates := make([][]rune, 0)
	allCommands := make([]string, 0)
	allCommands = append(allCommands, t.builtIns...)
	allCommands = append(allCommands, t.path_executables...)

	// retain only unique
	allCommands = getUniqueStrings(allCommands) // Ensure allCommands contains unique commands
	sort.Strings(allCommands)                   // Sort the commands for better output

	for _, cmd := range allCommands {
		if len(cmd) >= pos && string(line) == cmd[:pos] {
			candidates = append(candidates, []rune(cmd)) // Added space at the end
		}
	}

	if len(t.lastEnteredLine) > 0 && string(line) != t.lastEnteredLine {
		t.tabPressedAfterMultipleResults = false // Reset if the line has changed
	}
	t.lastEnteredLine = string(line)

	if len(candidates) == 0 {
		ringBell()
		t.tabPressedAfterMultipleResults = false
		return nil, pos
	}

	if len(candidates) == 1 {
		t.tabPressedAfterMultipleResults = false
		candidateWithSpace := append(candidates[0][pos:], ' ')
		// just return candidate[0] appended with space, but return as [][]rune
		return [][]rune{candidateWithSpace}, pos
	}

	if !t.tabPressedAfterMultipleResults {
		t.tabPressedAfterMultipleResults = true
		ringBell()
	} else {
		// print candidates separated by space
		fmt.Println()
		for i, candidate := range candidates {
			fmt.Print(string(candidate))
			if i < len(candidates)-1 { // Add space only for non-last elements
				fmt.Print("  ")
			}
		}
		fmt.Println() // Print a newline after all candidates
		fmt.Print("$ " + string(line))
		t.tabPressedAfterMultipleResults = false // Reset after displaying candidates
	}

	return nil, pos
}
