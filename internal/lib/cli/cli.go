package cli

import (
	"bufio"
	"fmt"
	"golang.org/x/term"
	"os"
	"strings"
)

func GetUserInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			fmt.Println("value can't be empty")
		} else {
			return input
		}
	}
}

func GetSensitiveUserInput(prompt string) (string, error) {
	for {
		fmt.Print(prompt)
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", err
		}

		fmt.Println() // Print a newline because ReadPassword does not capture the enter key

		if len(bytePassword) == 0 {
			fmt.Println("value can't be empty")
		} else {
			return string(bytePassword), nil
		}
	}
}
