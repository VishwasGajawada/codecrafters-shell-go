package trie

type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
	count    int
}

func NewTrieNode(words []string) *TrieNode {
	root := &TrieNode{}
	for _, word := range words {
		root.insert(word)
	}
	return root
}

func (t *TrieNode) insert(word string) {
	node := t
	for _, char := range word {
		if node.children == nil {
			node.children = make(map[rune]*TrieNode)
		}
		if _, exists := node.children[char]; !exists {
			node.children[char] = &TrieNode{}
		}
		node = node.children[char]
		node.count++ // count indicates how many words pass through this node
	}
	node.isEnd = true
}

func (t *TrieNode) GetAllMatching(prefix string) []string {
	node := t
	for _, char := range prefix {
		if node.children == nil || node.children[char] == nil {
			return nil // No matching prefix found
		}
		node = node.children[char]
	}

	return node.collectWords(prefix)
}

func (t *TrieNode) collectWords(prefix string) []string {
	var words []string
	if t.isEnd {
		words = append(words, prefix)
	}

	for char, child := range t.children {
		words = append(words, child.collectWords(prefix+string(char))...)
	}

	return words
}

func (t *TrieNode) GetLongestCommonPrefix(word string, limit int) string {
	node := t
	prefix := ""
	for _, char := range word {
		if node.children == nil || node.children[char] == nil {
			break
		}
		node = node.children[char]
		if node.count < limit {
			break // Stop if the count is less than the limit
		}
		prefix += string(char)
	}
	return prefix
}
