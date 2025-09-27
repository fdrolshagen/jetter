package parser

import (
	"bufio"
	"fmt"
	"github.com/fdrolshagen/jetter/internal"
	"io"
	"os"
	"strconv"
	"strings"
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
)

var methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

func ParseHttpFile(filename string) ([]internal.Request, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ParseHttp(file)
}

func ParseHttp(r io.Reader) ([]internal.Request, error) {
	var requests []internal.Request

	state := StateParsingStarted
	lineCounter := 0
	request := newRequest()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineCounter++
		//fmt.Printf("state: %d - line: \"%s\"\n", state, line)

		switch state {
		case StateParsingStarted:
			if strings.HasPrefix(line, "###") {
				request.Name = getName(line, len(requests)+1)
				state = StateInitialConfigLineRead
			}
		case StateInitialConfigLineRead:
			// read http line, parse method and url
			parts := strings.Fields(line)
			switch {
			case len(parts) == 1 && strings.HasPrefix(parts[0], "http"):
				request.Method = "GET"
				request.Url = parts[0]
			case len(parts) == 2 && isAllowedMethod(parts[0]):
				request.Method = parts[0]
				request.Url = parts[1]
			default:
				return nil, fmt.Errorf("parsing error: invalid request at line %d", lineCounter)
			}

			state = StateHttpConfigLineRead
		case StateHttpConfigLineRead, StateHttpHeaderRead:
			// either newline or header line following
			line = strings.TrimSpace(line)

			if line == "\n" || line == "" {
				state = StateHeaderBodySeparationRead
				break
			}

			if strings.HasPrefix(line, "###") {
				// new request starts now, append current request and reset request to start parsing again
				appendAndReset(&requests, &request)

				request.Name = getName(line, len(requests)+1)
				state = StateInitialConfigLineRead
				continue
			}

			if strings.HasPrefix(line, "#") {
				// ignore comments and duplicate initial config line
				continue
			}

			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("parsing error: invalid header at line %d", lineCounter)
			}
			request.Headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])

			state = StateHttpHeaderRead
		case StateHeaderBodySeparationRead, StateBodyPartRead, StateIgnoredBodyPartRead:
			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "###") {
				// new request starts now, append current request and reset request to start new parsing
				appendAndReset(&requests, &request)

				request.Name = getName(line, len(requests)+1)
				state = StateInitialConfigLineRead
				continue
			}

			if ignoreLine(line) {
				state = StateIgnoredBodyPartRead
				continue
			}

			request.Body += line + "\n"
			state = StateBodyPartRead
		default:
			return nil, fmt.Errorf("parsing error: invalid internal state")
		}

	}

	appendAndReset(&requests, &request)
	return requests, nil
}

func ignoreLine(line string) bool {
	// HasPrefix(line ">") -> javascript statement, ignore
	return line == "\n" || line == "" || strings.HasPrefix(line, ">")
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
