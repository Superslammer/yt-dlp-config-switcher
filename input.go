package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func readInputYN(failString string) bool {
	if failString == "" {
		failString = "Please only write y or n: "
	}

	for {
		input := bufio.NewScanner(os.Stdin)
		input.Scan()

		if strings.ToLower(input.Text()) == "n" {
			return false
		} else if strings.ToLower(input.Text()) == "y" {
			return true
		}

		fmt.Print("\n" + failString)
	}
}

func readInput(expectedStrings []string) string {
	for {
		fmt.Println()
		fmt.Print(">> ")

		input := bufio.NewScanner(os.Stdin)
		input.Scan()

		if expectedStrings != nil {
			for _, expected := range expectedStrings {
				if input.Text() == expected {
					return expected
				}
			}
		} else {
			return input.Text()
		}
	}
}
