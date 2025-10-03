package executor

import (
	"github.com/fdrolshagen/jetter/internal"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSubmit_ZeroDuration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	s := internal.Scenario{
		Duration: 0,
		Collection: &internal.Collection{
			Requests: []internal.Request{{Method: "GET", Url: server.URL}},
		},
	}
	result := Submit(s)
	assert.Len(t, result.Executions, 1)
	assert.False(t, result.AnyError)
}

func TestSubmit_WithDuration(t *testing.T) {
	s := internal.Scenario{
		Duration: 30 * time.Millisecond,
		Collection: &internal.Collection{
			Requests: []internal.Request{{Method: "GET", Url: "http://localhost"}},
		},
	}
	result := Submit(s)
	assert.GreaterOrEqual(t, len(result.Executions), 2)
}

func TestExecuteScenario_ErrorInvalidRequest(t *testing.T) {
	s := internal.Scenario{Collection: &internal.Collection{
		Requests: []internal.Request{{Method: "", Url: ""}},
	}}
	exec := ExecuteScenario(s)
	assert.True(t, exec.AnyError)
}

func TestExecuteScenario_ExecutesRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
	}))
	defer server.Close()

	s := internal.Scenario{
		Collection: &internal.Collection{
			Requests: []internal.Request{{Method: "GET", Url: server.URL}},
		},
	}
	exec := ExecuteScenario(s)
	assert.False(t, exec.AnyError)
	assert.Len(t, exec.Responses, 1)
	assert.Equal(t, 201, exec.Responses[0].Status)
}

func TestExecuteRequest_ErrorOnBadRequest(t *testing.T) {
	resp := ExecuteRequest(internal.Request{Method: "BAD", Url: ":://"})
	assert.NotNil(t, resp.Error)
}

func TestExecuteRequest_RealRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(202)
	}))
	defer server.Close()

	req := internal.Request{Method: "GET", Url: server.URL}
	resp := ExecuteRequest(req)
	assert.Nil(t, resp.Error)
	assert.Equal(t, 202, resp.Status)
	assert.GreaterOrEqual(t, int(resp.Duration), 0)
}
