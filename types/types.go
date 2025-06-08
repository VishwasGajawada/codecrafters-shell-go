package types

import "os"

// Command represents a parsed shell command.
type Command struct {
	Name         string
	Args         []string
	OutputStream *os.File
	ErrorStream  *os.File
	// Potentially add InputStream for '<' redirects later
}
