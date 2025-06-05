package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

func main() {
	// Print the shell prompt
	for {
		fmt.Fprint(os.Stdout, "$ ")
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			panic(err)
		}
		// if command is empty, break
		if strings.TrimSpace(command) == "" {
			break
		}
		// split the command by spaces into array of words
		words := strings.Fields(command)
		if words[0] == "exit" {
			return
		}

		if words[0] == "echo" {
			// Print the rest of the command
			if len(words) > 1 {
				fmt.Fprintln(os.Stdout, strings.Join(words[1:], " "))
			} else {
				fmt.Fprintln(os.Stdout, "")
			}
			continue
		}
		fmt.Fprintf(os.Stdout, "%s: command not found\n", strings.TrimSpace(command))
	}
}
