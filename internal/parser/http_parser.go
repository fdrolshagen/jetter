package parser

import (
	"bufio"
	"fmt"
	"github.com/fdrolshagen/jetter/internal"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type State int

const (
	StateParsingStarted State = iota
	StateInitialConfigLineRead
	StateHttpConfigLineRead
	StateHttpHeaderRead
	StateHeaderBodySeparationRead
	StateBodyPartRead
	StateIgnoredBodyPartRead
	StateMultilineScriptStarted
)

var methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

func ParseHttpFile(filename string) (internal.Collection, error) {
	file, err := os.Open(filename)
	if err != nil {
		return internal.Collection{}, err
	}
	defer file.Close()

	return ParseHttp(file)
}

func ParseHttp(r io.Reader) (internal.Collection, error) {
	var requests []internal.Request
	var vars = map[string]string{}

	state := StateParsingStarted
	lineCounter := 0
	request := newRequest()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		rawLine := scanner.Text()
		line := strings.TrimSpace(rawLine)
		lineCounter++

		if handleNewRequest(line, &requests, &request) {
			state = StateInitialConfigLineRead
			continue
		}

		switch state {
		case StateParsingStarted:
			if err := handleVariableDefinition(line, vars, lineCounter); err != nil {
				return internal.Collection{}, err
			}
		case StateInitialConfigLineRead:
			if err := handleRequestLine(line, &request, lineCounter); err != nil {
				return internal.Collection{}, err
			}
			state = StateHttpConfigLineRead
		case StateHttpConfigLineRead, StateHttpHeaderRead:
			if handled, err := handleJetterDirectiveLine(line, &request, lineCounter); handled {
				if err != nil {
					return internal.Collection{}, err
				}
				continue
			}
			if isHeaderBodySeparation(line) {
				state = StateHeaderBodySeparationRead
				break
			}
			if isComment(line) {
				continue
			}
			if err := handleHeaderLine(line, &request, lineCounter); err != nil {
				return internal.Collection{}, err
			}
			state = StateHttpHeaderRead
		case StateHeaderBodySeparationRead, StateBodyPartRead, StateIgnoredBodyPartRead:
			if handled, err := handleJetterDirectiveLine(line, &request, lineCounter); handled {
				if err != nil {
					return internal.Collection{}, err
				}
				continue
			}
			if isPostRequestScriptStart(line) {
				if isSingleLineScript(line) {
					appendScriptLine(&request, extractSingleLineScript(line))
				} else {
					appendScriptLine(&request, extractScriptStartLine(line))
					state = StateMultilineScriptStarted
				}
				continue
			}
			if isEmptyLine(line) || isScriptOrFile(line) {
				state = StateIgnoredBodyPartRead
				continue
			}
			request.Body += rawLine + "\n"
			state = StateBodyPartRead
		case StateMultilineScriptStarted:
			if isScriptEnd(line) {
				appendScriptLine(&request, extractScriptEndLine(line))
				state = StateIgnoredBodyPartRead
				continue
			}
			appendScriptLine(&request, line)
		default:
			return internal.Collection{}, fmt.Errorf("parsing error: invalid internal state")
		}
	}

	if err := scanner.Err(); err != nil {
		return internal.Collection{}, fmt.Errorf("parsing error: failed to read input: %w", err)
	}

	appendAndReset(&requests, &request)
	return internal.Collection{
		Requests:  requests,
		Variables: vars,
	}, nil
}

func isNewRequest(line string) bool {
	return strings.HasPrefix(line, "###")
}

func isVariableDefinition(line string) bool {
	return strings.HasPrefix(line, "@")
}

func isHeaderBodySeparation(line string) bool {
	return line == "\n" || line == ""
}

func isComment(line string) bool {
	return strings.HasPrefix(line, "#")
}

func isScriptOrFile(line string) bool {
	return strings.HasPrefix(line, ">") || strings.HasPrefix(line, "<")
}

func isJetterDirective(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "@jetter.") || strings.HasPrefix(trimmed, "#@jetter.")
}

func isPostRequestScriptStart(line string) bool {
	return strings.HasPrefix(line, "> {%")
}

func isScriptEnd(line string) bool {
	return strings.HasSuffix(line, "%}")
}

func isSingleLineScript(line string) bool {
	return isPostRequestScriptStart(line) && isScriptEnd(line)
}

func extractSingleLineScript(line string) string {
	line = strings.TrimPrefix(line, "> {%")
	line = strings.TrimSuffix(line, "%}")
	return strings.TrimSpace(line)
}

func extractScriptStartLine(line string) string {
	line = strings.TrimPrefix(line, "> {%")
	return strings.TrimSpace(line)
}

func extractScriptEndLine(line string) string {
	line = strings.TrimSuffix(line, "%}")
	return strings.TrimSpace(line)
}

func appendScriptLine(request *internal.Request, scriptLine string) {
	if scriptLine == "" {
		return
	}
	request.PostScript += scriptLine + "\n"
}

func handleJetterDirectiveLine(line string, request *internal.Request, lineCounter int) (bool, error) {
	if !isJetterDirective(line) {
		return false, nil
	}

	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return true, fmt.Errorf("parsing error: invalid @jetter directive at line %d", lineCounter)
	}

	key := strings.TrimSpace(parts[0])
	key = strings.TrimPrefix(key, "#")
	value := strings.TrimSpace(parts[1])

	switch key {
	case "@jetter.while":
		request.JetterWhile = value
	case "@jetter.maxIterations":
		maxIterations, err := strconv.Atoi(value)
		if err != nil || maxIterations <= 0 {
			return true, fmt.Errorf("parsing error: invalid @jetter.maxIterations at line %d", lineCounter)
		}
		request.JetterMaxIterations = maxIterations
	case "@jetter.sleep":
		sleep, err := time.ParseDuration(value)
		if err != nil || sleep < 0 {
			return true, fmt.Errorf("parsing error: invalid @jetter.sleep at line %d", lineCounter)
		}
		request.JetterSleep = sleep
	case "@jetter.onTimeout":
		onTimeout := strings.ToLower(value)
		if onTimeout != "fail" && onTimeout != "continue" {
			return true, fmt.Errorf("parsing error: invalid @jetter.onTimeout at line %d", lineCounter)
		}
		request.JetterOnTimeout = onTimeout
	default:
		return true, fmt.Errorf("parsing error: unknown @jetter directive at line %d", lineCounter)
	}

	return true, nil
}

func handleVariableDefinition(line string, vars map[string]string, lineCounter int) error {
	if !isVariableDefinition(line) {
		return nil
	}

	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("parsing error: invalid variable definition at line %d", lineCounter)
	}
	key := strings.TrimSpace(parts[0])
	key = strings.TrimLeft(key, "@")
	value := strings.TrimSpace(parts[1])
	vars[key] = value
	return nil
}

func handleNewRequest(line string, requests *[]internal.Request, request *internal.Request) bool {
	if isNewRequest(line) {
		appendAndReset(requests, request)
		request.Name = getName(line, len(*requests)+1)
		return true
	}
	return false
}

func handleRequestLine(line string, request *internal.Request, lineCounter int) error {
	parts := strings.Fields(line)
	if len(parts) == 1 && strings.HasPrefix(parts[0], "http") {
		request.Method = "GET"
		request.Url = parts[0]
	} else if len(parts) == 2 && isAllowedMethod(parts[0]) {
		request.Method = parts[0]
		request.Url = parts[1]
	} else {
		return fmt.Errorf("parsing error: invalid request at line %d", lineCounter)
	}
	return nil
}

func handleHeaderLine(line string, request *internal.Request, lineCounter int) error {
	if !strings.Contains(line, ":") && line != "" {
		return fmt.Errorf("parsing error: expected blank line between headers and body at line %d", lineCounter)
	}
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("parsing error: invalid header at line %d", lineCounter)
	}
	request.Headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	return nil
}

func isEmptyLine(line string) bool {
	return line == "\n" || line == ""
}

func newRequest() internal.Request {
	return internal.Request{
		Headers: map[string]string{},
	}
}

func appendAndReset(requests *[]internal.Request, request *internal.Request) {
	if request.Method != "" && request.Url != "" {
		*requests = append(*requests, *request)
	}
	*request = newRequest()
}

func getName(line string, counter int) string {
	name := strings.TrimLeft(line, "###")
	name = strings.TrimSpace(name)

	if name == "" {
		name = "Request #" + strconv.Itoa(counter)
	}

	return name
}

func isAllowedMethod(method string) bool {
	if method == "" {
		return false
	}

	for _, m := range methods {
		if m == method {
			return true
		}
	}

	return false
}
