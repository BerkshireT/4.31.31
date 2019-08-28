//**********************************************************
//
//    filename: diff1.go
//
//    description: Go implementation of linux diff
//
//    author: Tyler Berkshire
//    login id: FA_19_CPS444_02
//
//    class:  CPS 444
//    instructor:  Perugini
//    assignment:  Homework 1
//
//    assigned: Aug 22
//    due: Aug 29
//
//**********************************************************

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

func main() {
	flags := parseFlags()

	args := flag.Args()
	handleInput(args)

	var pathToFile string
	var pathToFile2 string
	var hasStandardIn bool

	// Check for standard input in place of a file
	if args[0] == "-" {
		pathToFile = args[1]
		hasStandardIn = true
	} else if args[1] == "-" {
		pathToFile = args[0]
		hasStandardIn = true
	} else {
		pathToFile = args[0]
		pathToFile2 = args[1]
	}

	var file1 []string
	var file2 []string

	// Copy the file, or stdin, to []string
	if hasStandardIn {
		file1 = readStandardInput()
		file2 = readFile(pathToFile)
	} else {
		file1 = readFile(pathToFile)
		file2 = readFile(pathToFile2)
	}

	// Trim files according to flags
	if flags[0] {
		file1 = trimLeading(file1)
		file2 = trimLeading(file2)
	}
	if flags[1] {
		file1 = trimTrailing(file1)
		file2 = trimTrailing(file2)
	}
	if flags[2] {
		file1 = trimMiddle(file1)
		file2 = trimMiddle(file2)
	}
	if flags[3] {
		file1 = trimAll(file1)
		file2 = trimAll(file2)
	}

	// Compare
	results := doExactComp(file1, file2)

	printResults(results)
	os.Exit(0)
}

func parseFlags() []bool {
	lead := flag.Bool("l", false, "ignore leading whitespace")
	trail := flag.Bool("t", false, "ignore trailing whitespace")
	middle := flag.Bool("m", false, "ignore intermediary whitespace")
	all := flag.Bool("a", false, "ignore all whitespace")

	flag.Parse()

	if (*all && *lead) || (*all && *trail) || (*all && *middle) {
		fmt.Fprintln(os.Stderr, "diff1: Option -a cannot be combined with any other options.")
		os.Exit(9)
	}

	return []bool{*lead, *trail, *middle, *all}
}

func handleInput(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "diff1: missing operand after 'diff.go'\n")
		os.Exit(2)
	} else if len(args) == 1 {
		fmt.Fprintf(os.Stderr, "diff1: missing operand after '%s'\n", args[0])
		os.Exit(2)
	} else if len(args) > 2 {
		fmt.Fprintf(os.Stderr, "diff1: extra operand '%s'\n", args[2])
		os.Exit(2)
	}

	invalidPath := false
	if args[0] != "-" && !doesFileExist(args[0]) {
		invalidPath = true // arg 1 path not valid
	}

	if args[1] != "-" && !doesFileExist(args[1]) {
		invalidPath = true // arg 2 path not valid
	}

	if invalidPath {
		os.Exit(2)
	}

	if args[0] == "-" && args[1] == "-" {
		readStandardInput()
		os.Exit(0) // stdin will always equal itself
	}
}

func doesFileExist(x string) bool {
	validInput := true

	if _, err := os.Stat(x); err != nil { // file doesn't exist
		fmt.Fprintf(os.Stderr, "diff1: %s: No such file or directory\n", x)
		validInput = false
	}

	return validInput
}

func readStandardInput() []string {
	input := make([]string, 0)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if err := scanner.Err(); err == io.EOF {
			break // scan until end of file
		} else {
			input = append(input, scanner.Text())
		}
	}
	return input
}

func trimLeading(before []string) []string {
	after := make([]string, 0)
	for i := 0; i < len(before); i++ {
		after = append(after, strings.TrimLeft(before[i], " "))
	}
	return after
}

func trimTrailing(before []string) []string {
	after := make([]string, 0)
	for i := 0; i < len(before); i++ {
		after = append(after, strings.TrimRight(before[i], " "))
	}
	return after
}

func trimMiddle(before []string) []string {
	after := make([]string, 0)
	for i := 0; i < len(before); i++ {
		var newLine string
		lineArray := []rune(before[i])

		noLead := strings.TrimLeft(before[i], " ")
		noTrail := strings.TrimRight(before[i], " ")

		// Find where the trailing/leading whitespace stops
		startString := strings.Index(before[i], string(noLead[0]))
		endString := strings.LastIndex(before[i], string(noTrail[len(noTrail)-1]))

		// Copy over leading whitespace
		for j := 0; j < startString; j++ {
			newLine += string(lineArray[j])
		}

		// Trim middle whitespace
		for j := startString; j <= endString; j++ {
			if string(lineArray[j]) != " " {
				newLine += string(lineArray[j])
			}
		}

		// Copy trailing whitespace
		for j := endString + 1; j < len(lineArray); j++ {
			newLine += string(lineArray[j])
		}

		after = append(after, newLine)
	}
	return after
}

func trimAll(before []string) []string {
	after := make([]string, 0)
	for i := 0; i < len(before); i++ {
		after = append(after, strings.Join(strings.Fields(before[i]), ""))
	}
	return after
}

func doExactComp(file1 []string, file2 []string) []int {
	var endIndex int
	var overEndIndex int

	// Determine where the larger file ends
	if len(file1) == len(file2) {
		endIndex = len(file1) - 1
	} else if len(file1) > len(file2) {
		endIndex = len(file2) - 1
		overEndIndex = len(file1)
	} else { // file 1 < file2
		endIndex = len(file1) - 1
		overEndIndex = len(file2)
	}

	// Make comparison
	results := make([]int, 0)
	for i := 0; i <= endIndex; i++ {
		if file1[i] != file2[i] {
			results = append(results, i+1)
		}
	}

	if len(file1) == len(file2) {
		return results // Same size means no extra lines
	}

	// Any extra lines will be different
	for i := endIndex + 2; i <= overEndIndex; i++ {
		results = append(results, i)
	}

	return results
}

func readFile(path string) []string {
	input := make([]string, 0)
	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if err := scanner.Err(); err == io.EOF {
			break // scan until end of file
		} else {
			input = append(input, scanner.Text())
		}
	}
	return input
}

func printResults(results []int) {
	sort.Ints(results)
	extra := make(map[int]bool)
	for _, line := range results {
		if isPresent := extra[line]; !isPresent {
			extra[line] = true
			fmt.Println(line)
		}
	}
}
