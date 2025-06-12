package shell

import (
	"fmt"
	"sort"

	"github.com/codecrafters-io/shell-starter-go/trie"
)

type TabCompleter struct {
	trie                           *trie.TrieNode // Assuming TrieNode is defined in trie package
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
	candidates := t.trie.GetAllMatching(string(line))

	// retain only unique
	candidates = getUniqueStrings(candidates) // Ensure candidates contains unique commands
	sort.Strings(candidates)                  // Sort the commands for better output

	if len(t.lastEnteredLine) > 0 && string(line) != t.lastEnteredLine {
		t.tabPressedAfterMultipleResults = false // Reset if the line has changed
	}
	t.lastEnteredLine = string(line)

	if len(candidates) == 0 {
		ringBell()
		return nil, pos
	}

	if len(candidates) == 1 {
		return [][]rune{append([]rune(candidates[0][pos:]), ' ')}, pos
	}

	// now there are multiple candidates
	longestPrefix := t.trie.GetLongestCommonPrefix(candidates[0], len(candidates))
	if len(longestPrefix) > pos {
		// If the longest common prefix is longer than the current position,
		// we can complete the line up to that prefix.
		return [][]rune{[]rune(longestPrefix)[pos:]}, len(longestPrefix)
	}

	if !t.tabPressedAfterMultipleResults {
		t.tabPressedAfterMultipleResults = true
		ringBell()
	} else {
		// print candidates separated by space
		fmt.Println()
		for i, candidate := range candidates {
			fmt.Print(string(candidate))
			if i < len(candidates)-1 {
				fmt.Print("  ")
			}
		}
		fmt.Println()
		fmt.Print("$ " + string(line))
		t.tabPressedAfterMultipleResults = false
	}

	return nil, pos
}
