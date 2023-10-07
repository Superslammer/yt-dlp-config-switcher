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
