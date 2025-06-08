package main

import (
	"github.com/codecrafters-io/shell-starter-go/shell" // Import the shell package
)

func main() {
	myShell := shell.NewShell() // Create a new shell instance
	myShell.Run()               // Run the shell's main loop
}
