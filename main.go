package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	matcherRegexInitialNumbers = regexp.MustCompile(`var [Aa-z0-9]{64}=[0-9]+`)
	matcherRegexMathBasic      = regexp.MustCompile(`[a-z0-9]{64}=(~|\^|\||&|[A-z0-9]{64})`)
	weirdFuncEndingRegex       = regexp.MustCompile(`}\([a-z0-9]{64},[a-z0-9]{64},[a-z0-9]{64}\)`)
)

func parseScript(script string) string {
	out := ""
	startingWeirdFunc := false
	startingWeirdFuncAnswer := ""
	startingWeirdMathOperation := false
	startingWeirdMathOperationAnswer := ""
	scriptParts := strings.Split(strings.Split(script, "\r\n")[1], ";")
	answers := make(map[string]int, 4)
	for _, part := range scriptParts {
		// this gets the initial numbers
		if matcherRegexInitialNumbers.MatchString(part) {
			matched := matcherRegexInitialNumbers.FindString(part)
			if matched == "" {
				panic("how???")
			}
			matchedParts := strings.Split(matched, "=")
			intValue, err := strconv.Atoi(matchedParts[1])
			if err != nil {
				panic(err)
			}
			func() {
				answers[strings.Split(matchedParts[0], " ")[1]] = intValue
				//fmt.Println(strings.Split(matchedParts[0], " ")[1], intValue)
			}()
		}
		// if it's a basic math operation
		if matcherRegexMathBasic.MatchString(part) && !strings.Contains(part, "new Date") {
			signChange := false
			operationDone := false
			// these are usually cancer so we gotta do some special shit for those :)
			if strings.Contains(part, "~") {
				if strings.Contains(part, "(") {
					a := strings.Split(part, "=")[1]
					a = a[2 : len(a)-1]
					lmao := strings.Split(part, "=")
					lmao[1] = a
					part = strings.Join(lmao, "=")
				} else {
					a := strings.Split(part, "=")[1]
					a = a[1:]
					lmao := strings.Split(part, "=")
					lmao[1] = a
					part = strings.Join(lmao, "=")
				}
				signChange = true
			}
			parts := strings.Split(part, "=")
			if strings.Contains(parts[1], "^") {
				m := strings.Split(parts[1], "^")
				answers[parts[0]] = answers[m[0]] ^ answers[m[1]]
				operationDone = true
			}
			if strings.Contains(parts[1], "|") {
				m := strings.Split(parts[1], "|")
				answers[parts[0]] = answers[m[0]] | answers[m[1]]
				operationDone = true
			}
			if strings.Contains(parts[1], "&") {
				m := strings.Split(parts[1], "&")
				answers[parts[0]] = answers[m[0]] & answers[m[1]]
				operationDone = true
			}
			if signChange {
				if operationDone {
					answers[parts[0]] = -(answers[parts[0]] + 1)
				} else {
					answers[parts[0]] = -(answers[parts[1]] + 1)
				}
			}
		}
		// date
		if strings.Contains(part, "new Date") {
			parts := strings.Split(part, "=")
			operationParts := strings.Split(parts[1], "^")
			answers[parts[0]] = answers[operationParts[0]] ^ time.UnixMilli(int64(answers[strings.Split(strings.Split(operationParts[1], "*")[0], "(")[1]]*10000000000)).UTC().Day()
		}
		// starting of that long div adding function cancer
		if strings.Contains(part, "document.createElement('div')") && !startingWeirdFunc {
			startingWeirdFunc = true
			startingWeirdFuncAnswer = strings.Split(part, "=function")[0]
		}
		// ending of that
		if weirdFuncEndingRegex.MatchString(part) && startingWeirdFunc {
			startingWeirdFunc = false
			things := strings.Split(part[2:len(part)-1], ",")
			answers[startingWeirdFuncAnswer] = weirdFunc1(answers[things[0]], answers[things[1]], answers[things[2]])
			startingWeirdFuncAnswer = ""
		}
		// starting of weird math operation thingy
		if strings.Contains(part, "function(){return this.") && !startingWeirdMathOperation {
			startingWeirdMathOperation = true
			startingWeirdMathOperationAnswer = strings.Split(part, "=")[0]
		}
		// ending of that
		if weirdFuncEndingRegex.MatchString(part) && startingWeirdMathOperation {
			startingWeirdMathOperation = false
			//third ^ fourth | (third ^ second)
			// middle ^ first | (middle ^ last)
			things := strings.Split(part[2:len(part)-1], ",")
			answers[startingWeirdMathOperationAnswer] = answers[things[1]] ^ answers[things[0]] | (answers[things[1]] ^ answers[things[2]])
			startingWeirdMathOperationAnswer = ""
		}
		// get the final output
		if strings.HasPrefix(part, "return {'rf") {
			out = strings.ReplaceAll(part, "'", `"`)
			out = strings.Split(out, " ")[1]
			for a, b := range answers {
				out = strings.ReplaceAll(out, ":"+a, ":"+fmt.Sprint(b))
			}
			break
		}
	}
	return out
}

func main() {
	f, _ := os.ReadFile("copy.js")
	fmt.Println(parseScript(string(f)))
}

// This shit took me far too long but it's the aids part that does shit with divs and whatnot, looks really aids but is actually doing something really simple.
func weirdFunc1(third, second, fourth int) int {
	num := 0
	for i := 0; i < 8; i++ {
		if (third & 1) == 0 {
			num += third
		}
		if (second & 1) == 0 {
			num += second
		}
		if (fourth & 1) == 0 {
			num += fourth
		}
		third = third >> 1
		second = second >> 1
		fourth = fourth >> 1
	}
	return num % 256
}
