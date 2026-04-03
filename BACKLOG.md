# Feature Backlog

---

### Report Format
- Add reporters for **JSON** and **YAML** output.
- Consider using an interface for multiple reporters: `func Report(r internal.Result)`.

- `--output [-o]`  
  - Support `-o` syntax similar to `kubectl`.
  - Output file/format. Default is human-friendly tabular output.  
      Machine-readable options should be available, e.g., `-o json` or `-o yaml`.

### Jetter-Specific Directives (Advanced)
- Support additional global configuration at the top of a `.http` file (beyond current loop directives):  
  `#@jetter threshold_http_req_failed 0.01`

### IntelliJ Request Configuration Support
- IntelliJ `.http` syntax allows using directives like `# @timeout 10`.
- Not all directives may be supported or relevant; this needs verification.
- Parser should handle supported directives, and the HTTP client should pick up the configuration.

### Token Refresh After Expiry
- If a token expires during a scenario, Jetter should automatically refresh or obtain a new token.
