# Feature Backlog

---

## High Priority

### Executor Support (Concurrency)
- Executor should support execution on multiple goroutines.
- Execution results should be returned from each goroutine after completion.
- `--concurrency [-c]` Number of threads to use for firing requests.

---

## Medium Priority

### Report Format
- Add reporters for **JSON** and **YAML** output.
- Consider using an interface for multiple reporters: `func Report(r internal.Result)`.

- `--output [-o]`  
  - Support `-o` syntax similar to `kubectl`.
  - Output file/format. Default is human-friendly tabular output.  
      Machine-readable options should be available, e.g., `-o json` or `-o yaml`.

### Jetter-Specific Directives (Per-Request or Global)
- Support global configuration at the top of a `.http` file:  
  `#@jetter threshold_http_req_failed 0.01`
- Support per-request configuration:  
  `#@jetter extract ID $.username`  
  Variables can then be reused in other requests: `{{$vars("ID")}}`

---

## Low Priority / Optional

### IntelliJ Request Configuration Support
- IntelliJ `.http` syntax allows using directives like `# @timeout 10`.
- Not all directives may be supported or relevant; this needs verification.
- Parser should handle supported directives, and the HTTP client should pick up the configuration.

### Token Refresh After Expiry
- If a token expires during a scenario, Jetter should automatically refresh or obtain a new token.
