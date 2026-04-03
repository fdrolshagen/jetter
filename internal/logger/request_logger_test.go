package logger

import (
	"encoding/json"
	"github.com/fdrolshagen/jetter/internal"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRequestLogger_WritesLogEntry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logs", "requests.log")
	requestLogger, err := NewRequestLogger(path)
	assert.NoError(t, err)
	defer requestLogger.Close()

	response := internal.Response{
		Name:           "GET users",
		RequestMethod:  "GET",
		RequestURL:     "http://localhost/users",
		RequestHeaders: map[string]string{"Authorization": "Bearer token"},
		RequestBody:    "",
		StartedAt:      time.Now().Add(-5 * time.Millisecond),
		Status:         200,
		Headers:        map[string]string{"Content-Type": "application/json"},
		Body:           `{"ok":true}`,
		Duration:       5 * time.Millisecond,
		FinishedAt:     time.Now(),
	}

	requestLogger.Log(response)
	assert.NoError(t, requestLogger.Err())

	content, err := os.ReadFile(path)
	assert.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.Len(t, lines, 1)

	var entry LogEntry
	err = json.Unmarshal([]byte(lines[0]), &entry)
	assert.NoError(t, err)
	assert.Equal(t, "GET", entry.Request.Method)
	assert.Equal(t, "http://localhost/users", entry.Request.URL)
	assert.Equal(t, 200, entry.Response.Status)
	assert.Equal(t, int64(5), entry.Response.DurationMs)
}
