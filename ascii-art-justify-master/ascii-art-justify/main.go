package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const charHeight = 8

var ansiColors = map[string]string{
	"red":    "\033[31m",
	"green":  "\033[32m",
	"yellow": "\033[33m",
	"blue":   "\033[34m",
	"purple": "\033[35m",
	"cyan":   "\033[36m",
	"white":  "\033[37m",
	"orange": "\033[38;5;208m",
	"reset":  "\033[0m",
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	var outputFileName, colorFlag, colorSubstr, reverseFile, alignType string
	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--output=") {
			outputFileName = strings.TrimPrefix(args[i], "--output=")
			args = append(args[:i], args[i+1:]...)
			i--
		} else if strings.HasPrefix(args[i], "--reverse=") {
			reverseFile = strings.TrimPrefix(args[i], "--reverse=")
			args = append(args[:i], args[i+1:]...)
			i--
		} else if strings.HasPrefix(args[i], "--align=") {
			alignType = strings.TrimPrefix(args[i], "--align=")
			args = append(args[:i], args[i+1:]...)
			i--
		} else if strings.HasPrefix(args[i], "--color=") {
			colorFlag = strings.TrimPrefix(args[i], "--color=")
			if i+1 < len(args) && !isBanner(args[i+1]) && !strings.HasPrefix(args[i+1], "--") {
				colorSubstr = args[i+1]
				args = append(args[:i], args[i+2:]...)
			} else {
				args = append(args[:i], args[i+1:]...)
			}
			i--
		}
	}

	if reverseFile != "" {
		reverseASCII(reverseFile)
		return
	}

	if len(args) == 0 {
		printUsage()
		return
	}

	input := args[0]
	banner := "standard"
	if len(args) > 1 {
		banner = args[1]
	}

	asciiMap, err := loadASCIITable(banner + ".txt")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	result := generateASCII(input, asciiMap, colorFlag, colorSubstr, alignType)

	if outputFileName != "" {
		os.WriteFile(outputFileName, []byte(result), 0644)
	} else {
		fmt.Print(result)
	}
}

func getTerminalWidth() int {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 80
	}
	parts := strings.Fields(string(out))
	if len(parts) < 2 {
		return 80
	}
	width, _ := strconv.Atoi(parts[1])
	return width
}

func isBanner(arg string) bool {
	return arg == "standard" || arg == "shadow" || arg == "thinkertoy"
}

func printUsage() {
	fmt.Println("Usage: go run . [OPTION] [STRING] [BANNER]")
	fmt.Println("EX: go run . --align=right --color=red \"hello\" standard")
}

func loadASCIITable(filename string) (map[rune][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	asciiMap := make(map[rune][]string)
	currentChar := rune(32)
	for i := 0; i < len(lines); i++ {
		if lines[i] == "" && i+8 < len(lines) {
			asciiMap[currentChar] = lines[i+1 : i+9]
			currentChar++
			i += 8
		}
	}
	return asciiMap, nil
}

func generateASCII(input string, asciiMap map[rune][]string, colorName, subStr, align string) string {
	if input == "" { return "" }
	input = strings.ReplaceAll(input, "\\n", "\n")
	parts := strings.Split(input, "\n")
	termWidth := getTerminalWidth()

	var builder strings.Builder
	colorCode := ansiColors[strings.ToLower(colorName)]
	resetCode := ansiColors["reset"]

	for _, part := range parts {
		if part == "" {
			builder.WriteString("\n")
			continue
		}

		totalContentWidth := 0
		for _, char := range part {
			if art, ok := asciiMap[char]; ok {
				totalContentWidth += len(art[0])
			}
		}

		words := strings.Fields(part)
		for row := 0; row < charHeight; row++ {
			line := ""
			switch align {
			case "center":
				padding := (termWidth - totalContentWidth) / 2
				if padding > 0 { line += strings.Repeat(" ", padding) }
			case "right":
				padding := termWidth - totalContentWidth
				if padding > 0 { line += strings.Repeat(" ", padding) }
			case "justify":
				if len(words) > 1 {
					totalWordsWidth := 0
					for _, word := range words {
						for _, char := range word {
							totalWordsWidth += len(asciiMap[char][0])
						}
					}
					spacesNeeded := termWidth - totalWordsWidth
					spaceWidth := spacesNeeded / (len(words) - 1)
					
					for i, word := range words {
						for _, char := range word {
							line += getColoredChar(char, row, asciiMap, colorCode, resetCode, subStr, part)
						}
						if i < len(words)-1 {
							line += strings.Repeat(" ", spaceWidth)
						}
					}
					builder.WriteString(line + "\n")
					continue
				}
			}

			if align != "justify" || len(words) <= 1 {
				for _, char := range part {
					line += getColoredChar(char, row, asciiMap, colorCode, resetCode, subStr, part)
				}
			}
			builder.WriteString(line + "\n")
		}
	}
	return builder.String()
}

func getColoredChar(char rune, row int, asciiMap map[rune][]string, colorCode, resetCode, subStr, part string) string {
	art := asciiMap[char][row]
	if colorCode != "" {
		if subStr == "" || strings.Contains(part, subStr) {
			if subStr == "" || strings.Contains(string(char), subStr) {
				return colorCode + art + resetCode
			}
		}
	}
	return art
}

func reverseASCII(fileName string) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	artLines := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
	if len(artLines) > 0 && artLines[len(artLines)-1] == "" {
		artLines = artLines[:len(artLines)-1]
	}
	refMap, _ := loadASCIITable("standard.txt")
	var result strings.Builder
	currentPos := 0
	maxW := 0
	for _, l := range artLines {
		if len(l) > maxW { maxW = len(l) }
	}
	for currentPos < maxW {
		found := false
		for char := rune(32); char <= rune(126); char++ {
			pattern, ok := refMap[char]
			if !ok { continue }
			width := len(pattern[0])
			if currentPos+width > maxW { continue }
			match := true
			for r := 0; r < charHeight; r++ {
				actualLine := ""
				if r < len(artLines) && currentPos < len(artLines[r]) {
					end := currentPos + width
					if end > len(artLines[r]) { end = len(artLines[r]) }
					actualLine = artLines[r][currentPos:end]
				}
				if len(actualLine) < width { actualLine += strings.Repeat(" ", width-len(actualLine)) }
				if actualLine != pattern[r] {
					match = false
					break
				}
			}
			if match {
				result.WriteRune(char)
				currentPos += width
				found = true
				break
			}
		}
		if !found { currentPos++ }
	}
	fmt.Println(result.String())
}