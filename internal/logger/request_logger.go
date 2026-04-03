package logger

import (
	"encoding/json"
	"github.com/fdrolshagen/jetter/internal"
	"os"
	"path/filepath"
	"sync"
)

type RequestLogger struct {
	mu       sync.Mutex
	file     *os.File
	writeErr error
}

type LogEntry struct {
	Request struct {
		Name      string            `json:"name"`
		Method    string            `json:"method"`
		URL       string            `json:"url"`
		Headers   map[string]string `json:"headers"`
		Body      string            `json:"body"`
		StartedAt string            `json:"startedAt"`
	} `json:"request"`
	Response struct {
		Status     int               `json:"status"`
		Headers    map[string]string `json:"headers"`
		Body       string            `json:"body"`
		DurationMs int64             `json:"durationMs"`
		Error      string            `json:"error,omitempty"`
		FinishedAt string            `json:"finishedAt"`
	} `json:"response"`
}

func NewRequestLogger(path string) (*RequestLogger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return &RequestLogger{file: file}, nil
}

func (l *RequestLogger) Log(response internal.Response) {
	if l == nil {
		return
	}

	entry := toLogEntry(response)
	encoded, err := json.Marshal(entry)
	if err != nil {
		l.setWriteErr(err)
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.writeErr != nil {
		return
	}

	if _, err := l.file.Write(append(encoded, '\n')); err != nil {
		l.writeErr = err
	}
}

func (l *RequestLogger) Err() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.writeErr
}

func (l *RequestLogger) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	return l.file.Close()
}

func (l *RequestLogger) setWriteErr(err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.writeErr == nil {
		l.writeErr = err
	}
}

func toLogEntry(response internal.Response) LogEntry {
	entry := LogEntry{}
	entry.Request.Name = response.Name
	entry.Request.Method = response.RequestMethod
	entry.Request.URL = response.RequestURL
	entry.Request.Headers = response.RequestHeaders
	entry.Request.Body = response.RequestBody
	entry.Request.StartedAt = response.StartedAt.Format("2006-01-02T15:04:05.000000000Z07:00")

	entry.Response.Status = response.Status
	entry.Response.Headers = response.Headers
	entry.Response.Body = response.Body
	entry.Response.DurationMs = response.Duration.Milliseconds()
	entry.Response.FinishedAt = response.FinishedAt.Format("2006-01-02T15:04:05.000000000Z07:00")
	if response.Error != nil {
		entry.Response.Error = response.Error.Error()
	}

	return entry
}
