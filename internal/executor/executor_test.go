package executor

import (
	"context"
	"github.com/fdrolshagen/jetter/internal"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	s := internal.Scenario{
		Duration: 30 * time.Millisecond,
		Collection: &internal.Collection{
			Requests: []internal.Request{{Method: "GET", Url: server.URL}},
		},
	}
	result := Submit(s)
	assert.GreaterOrEqual(t, len(result.Executions), 1)
}

func TestSubmit_WithConcurrency(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	s := internal.Scenario{
		Duration:    30 * time.Millisecond,
		Concurrency: 2,
		Collection: &internal.Collection{
			Requests: []internal.Request{{Method: "GET", Url: server.URL}},
		},
	}
	result := Submit(s)
	assert.GreaterOrEqual(t, len(result.Executions), 2)
	assert.False(t, result.AnyError)
}

func TestSubmit_WithNonPositiveConcurrencyDefaultsToOneWorker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	s := internal.Scenario{
		Duration:    25 * time.Millisecond,
		Concurrency: 0,
		Collection: &internal.Collection{
			Requests: []internal.Request{{Method: "GET", Url: server.URL}},
		},
	}

	result := Submit(s)
	assert.NotEmpty(t, result.Executions)
	assert.False(t, result.AnyError)
}

func TestExecuteScenario_ErrorInvalidRequest(t *testing.T) {
	s := internal.Scenario{Collection: &internal.Collection{
		Requests: []internal.Request{{Method: "", Url: ""}},
	}}
	exec := ExecuteScenario(context.Background(), s)
	assert.True(t, exec.AnyError)
}

func TestExecuteScenario_ErrorWhenVariableEvaluationFails(t *testing.T) {
	s := internal.Scenario{Collection: &internal.Collection{
		Variables: map[string]string{"ID": "{{$random.unknown(1)}}"},
		Requests:  []internal.Request{{Method: "GET", Url: "http://localhost/users/{{ID}}"}},
	}}

	exec := ExecuteScenario(context.Background(), s)
	assert.True(t, exec.AnyError)
	assert.Nil(t, exec.Responses)
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
	exec := ExecuteScenario(context.Background(), s)
	assert.False(t, exec.AnyError)
	assert.Len(t, exec.Responses, 1)
	assert.Equal(t, 201, exec.Responses[0].Status)
}

func TestExecuteScenario_AppliesPostScriptVariablesToFollowingRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"token":"abc123"}`))
		case "/users":
			if r.Header.Get("Authorization") != "Bearer abc123" {
				w.WriteHeader(401)
				return
			}
			w.WriteHeader(200)
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	s := internal.Scenario{
		Collection: &internal.Collection{
			Requests: []internal.Request{
				{
					Method:     "GET",
					Url:        server.URL + "/token",
					PostScript: `client.global.set("TOKEN", response.body.token)`,
				},
				{
					Method: "GET",
					Url:    server.URL + "/users",
					Headers: map[string]string{
						"Authorization": "Bearer {{TOKEN}}",
					},
				},
			},
		},
	}

	exec := ExecuteScenario(context.Background(), s)
	assert.False(t, exec.AnyError)
	assert.Len(t, exec.Responses, 2)
	assert.Equal(t, 200, exec.Responses[1].Status)
}

func TestExecuteScenario_MarksExecutionAsFailedWhenPostScriptFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"token":"abc123"}`))
	}))
	defer server.Close()

	s := internal.Scenario{
		Collection: &internal.Collection{
			Requests: []internal.Request{
				{
					Method:     "GET",
					Url:        server.URL,
					PostScript: `client.global.set("TOKEN"`,
				},
			},
		},
	}

	exec := ExecuteScenario(context.Background(), s)
	assert.True(t, exec.AnyError)
	assert.Len(t, exec.Responses, 1)
	assert.Error(t, exec.Responses[0].Error)
	assert.True(t, strings.Contains(exec.Responses[0].Error.Error(), "post-script execution failed"))
}

func TestExecuteScenario_JetterWhileLoopsUntilConditionIsFalse(t *testing.T) {
	var attempts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Content-Type", "application/json")
		if attempts < 3 {
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"status":"PENDING"}`))
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status":"DONE"}`))
	}))
	defer server.Close()

	s := internal.Scenario{Collection: &internal.Collection{Requests: []internal.Request{{
		Method:              "GET",
		Url:                 server.URL,
		JetterWhile:         `response.body.status != "DONE"`,
		JetterMaxIterations: 30,
		JetterSleep:         1 * time.Millisecond,
		JetterOnTimeout:     "fail",
	}}}}

	exec := ExecuteScenario(context.Background(), s)
	assert.False(t, exec.AnyError)
	assert.Len(t, exec.Responses, 3)
	assert.Equal(t, 3, attempts)
	assert.Contains(t, exec.Responses[2].Body, "DONE")
}

func TestExecuteScenario_JetterWhileTimeoutFailsWhenConfigured(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status":"PENDING"}`))
	}))
	defer server.Close()

	s := internal.Scenario{Collection: &internal.Collection{Requests: []internal.Request{{
		Method:              "GET",
		Url:                 server.URL,
		JetterWhile:         `response.body.status != "DONE"`,
		JetterMaxIterations: 2,
		JetterOnTimeout:     "fail",
	}}}}

	exec := ExecuteScenario(context.Background(), s)
	assert.True(t, exec.AnyError)
	assert.Len(t, exec.Responses, 2)
	assert.Error(t, exec.Responses[1].Error)
	assert.Contains(t, exec.Responses[1].Error.Error(), "still true")
}

func TestExecuteScenario_JetterWhileTimeoutContinuesWhenConfigured(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status":"PENDING"}`))
	}))
	defer server.Close()

	s := internal.Scenario{Collection: &internal.Collection{Requests: []internal.Request{{
		Method:              "GET",
		Url:                 server.URL,
		JetterWhile:         `response.body.status != "DONE"`,
		JetterMaxIterations: 2,
		JetterOnTimeout:     "continue",
	}}}}

	exec := ExecuteScenario(context.Background(), s)
	assert.False(t, exec.AnyError)
	assert.Len(t, exec.Responses, 2)
	assert.NoError(t, exec.Responses[1].Error)
}

func TestExecuteRequest_ErrorOnBadRequest(t *testing.T) {
	resp := ExecuteRequest(context.Background(), internal.Request{Method: "BAD", Url: ":://"})
	assert.NotNil(t, resp.Error)
}

func TestExecuteRequest_RealRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(202)
	}))
	defer server.Close()

	req := internal.Request{Method: "GET", Url: server.URL}
	resp := ExecuteRequest(context.Background(), req)
	assert.Nil(t, resp.Error)
	assert.Equal(t, 202, resp.Status)
	assert.GreaterOrEqual(t, int(resp.Duration), 0)
}

func TestExecuteRequest_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	resp := ExecuteRequest(ctx, internal.Request{Method: "GET", Url: server.URL})
	assert.Error(t, resp.Error)
	assert.Contains(t, resp.Error.Error(), "context canceled")
}

func TestExecuteRequest_ContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	resp := ExecuteRequest(ctx, internal.Request{Method: "GET", Url: server.URL})
	assert.Error(t, resp.Error)
	assert.Contains(t, resp.Error.Error(), "context deadline exceeded")
}

func TestWithDefaultTimeout_AddsDeadlineWhenMissing(t *testing.T) {
	ctx, cancel := withDefaultTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, ok := ctx.Deadline()
	assert.True(t, ok)
}

func TestWithDefaultTimeout_PreservesExistingDeadline(t *testing.T) {
	baseCtx, baseCancel := context.WithTimeout(context.Background(), time.Second)
	defer baseCancel()

	ctx, cancel := withDefaultTimeout(baseCtx, 100*time.Millisecond)
	defer cancel()

	baseDeadline, baseOK := baseCtx.Deadline()
	deadline, ok := ctx.Deadline()
	assert.True(t, baseOK)
	assert.True(t, ok)
	assert.Equal(t, baseDeadline, deadline)
}
