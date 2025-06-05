package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

// func main() {
// 	// Uncomment this block to pass the first stage
// 	fmt.Fprint(os.Stdout, "$ ")

// 	// Wait for user input
// 	command, err := bufio.NewReader(os.Stdin).ReadString('\n')
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Fprintf(os.Stdout, "%s: command not found", command[:len(command)-1])
// }

func main() {
	// Print the shell prompt
	fmt.Fprint(os.Stdout, "$ ")

	// Wait for user input
	command, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		panic(err)
	}

	// Trim whitespace and handle empty input
	command = strings.TrimSpace(command)
	// Print the error message
	fmt.Fprintf(os.Stdout, "%s: command not found\n", command)
}
