# 🚀 Jetter

[![build & test](https://github.com/fdrolshagen/jetter/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/fdrolshagen/jetter/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/fdrolshagen/jetter)](https://goreportcard.com/report/github.com/fdrolshagen/jetter)
![GitHub Release](https://img.shields.io/github/v/release/fdrolshagen/jetter?include_prereleases&sort=semver&display_name=release)


**Jetter** is a light-weight load testing and API scenario runner for HTTP services. It uses a subset of the IntelliJ `.http` file syntax and supports environment variables, authentication, and more.

✨ **Why Jetter?**  
Unlike many other CLI-based HTTP testing tools, **Jetter** adheres to the [IntelliJ HTTP Client specification](https://www.jetbrains.com/help/idea/http-client-in-product-code-editor.html).  
That means you can:
- ✅ Write and run requests directly in `.http` files
- ✅ Use IntelliJ’s built-in syntax highlighting and auto-completion
- ✅ Seamlessly switch between IntelliJ and the CLI without changing formats

With Jetter, your `.http` files become reusable across development, testing, and load simulation — all while staying compatible with IntelliJ’s editor.

---

## Features

- **Parse and run `.http` scenario files**  
  Use the familiar HTTP request format from IntelliJ’s built-in HTTP client.

- **Seamless IntelliJ compatibility** ([specification](https://www.jetbrains.com/help/idea/http-client-in-product-code-editor.html))  
  Run the same requests you use in IntelliJ — with syntax highlighting, variables, and environment support.

  **✅ Supported**
  - 🌍 Environment variables
  - 📝 In-place and dynamic variables
  - 📑 Multiple requests per file
  - 🔑 OAuth2 authentication
  - 📂 Request body file input (`< path/to/body.json`)
  - 🧪 Post-request JavaScript (`> {% ... %}`)
  - 🔁 Jetter loop directives (`@jetter.*` / `#@jetter.*`)

  **🚫 Not yet supported**
  - 📤 Response output to file (`> file.txt`)
  - 🧩 Full IntelliJ HTTP JS runtime parity

- **Flexible execution modes**
  - ▶️ Run once for quick checks
  - 🔁 Execute continuously for a fixed duration
  - ⚙️ Simulate concurrency with multiple workers

- **Live and persisted observability**
  - 📊 Live-updating terminal table while requests run
  - 🧾 Optional request/response logging to file (`--log`)

---

## Quick Start

Installation via script:

```sh
curl -fsSL https://raw.githubusercontent.com/fdrolshagen/jetter/main/scripts/install.sh | bash
```

Run your first scenario with a single command:

```sh
jetter --file examples/example.http --env examples/http-client.env.json:local
```

---

## Command Line Flags

| Flag            | Alias | Description                                                 |
|-----------------|-------|-------------------------------------------------------------|
| `--help`        | `-h`  | Print available command line arguments and explanation      |
| `--file`        | `-f`  | Path to the .http file (required)                           |
| `--env`         | `-e`  | Path to the environment file. Format: `-e <file>:<env-key>` |
| `--duration`    | `-d`  | How long should the load test run (e.g. `30s`, `1m`)        |
| `--concurrency` | `-c`  | How many workers should run concurrently (default: 1)       |
| `--log`         |       | Path to newline-delimited JSON request logs                 |
| `--version`     |       | Print version and exit                                      |

---

## Example .http File

```text
@tsid = 0{{$random.hexadecimal(12)}}

### GET /users - get all users
GET {{URL}}/users
Authorization: Bearer {{$auth.token("auth-id")}}

### POST /users - create user
POST {{URL}}/users
Authorization: Bearer {{$auth.token("auth-id")}}
Content-Type: application/json

< requests/user.json

### GET /users/{tsid} - get a user by id
GET {{URL}}/users/{{tsid}}
Authorization: Bearer {{$auth.token("auth-id")}}

> {% client.global.set("NEW_TSID", response.body.id) %}

### GET /status - wait until processing is SUCCESS
GET {{URL}}/status
Authorization: Bearer {{$auth.token("auth-id")}}
#@jetter.while = response.body.status != "SUCCESS"
#@jetter.maxIterations = 10
#@jetter.sleep = 2s
#@jetter.onTimeout = fail

### DELETE /users/{tsid} - delete a user by id
DELETE {{URL}}/users/{{NEW_TSID}}
Authorization: Bearer {{$auth.token("auth-id")}}
```

---

## Example Environment File

```json
{
  "local": {
    "URL": "http://localhost:8080",
    "Security": {
      "Auth": {
        "auth-id": {
          "Type": "OAuth2",
          "Token URL": "http://localhost:8081/realms/test-realm/protocol/openid-connect/token",
          "Grant Type": "Password",
          "Client ID": "test-client",
          "Client Secret": "test-secret",
          "Username": "test-user",
          "Password": "test-password"
        }
      }
    }
  }
}
```

---

## In-place Variables

You can define **[in-place variables](https://www.jetbrains.com/help/idea/http-client-variables.html#in-place-variables)** directly at the top of your `.http` file using the `@` syntax.  

- Inline variables can be used in URLs, headers, and request bodies.
- Environment variables (from `--env`) are also available, but **inline variables take precedence** if keys overlap.

**Usage**

```text
@ID = 123
@TOKEN = abc

### Get User
GET http://localhost:8081/users/{{ID}}
Authorization: Bearer {{TOKEN}}
```


---

## Dynamic Variables

You can use built-in **[dynamic variables](https://www.jetbrains.com/help/idea/http-client-variables.html#dynamic-variables)** in your `.http` files.  

Jetter currently supports the following dynamic variables:

| Variable                     | Description                                       |
|------------------------------|---------------------------------------------------|
| `{{$random.uuid}}`  | Generates a random UUIDv4                         |
| `{{$random.hexadecimal(n)}}` | Generates a random hexadecimal string of length n |

**Usage**

```text
@UUID = {{$random.uuid}}
@TSID = 0{{$random.hexadecimal(12)}}

### Get User
GET http://localhost:8081/users/{{UUID}}

### Get User
GET http://localhost:8081/users/{{TSID}}
```

---

## OAuth 2.0 authorization
Jetter supports **[Oauth2 authentication](https://www.jetbrains.com/help/idea/oauth-2-0-authorization.html)** out of the box. You can define multiple auth configurations in your environment file and reference them in your `.http` file using the `{{$auth.token("auth-id")}}` magic variable. Supported Grant Types: `Client Credentials` and `Password`.

---

## Request body from file

You can define request bodies in separate files and reference them from `.http` scenarios using IntelliJ-style file input syntax:

```text
### Create User
POST {{URL}}/users
Authorization: Bearer {{$auth.token("auth-id")}}
Content-Type: application/json

< requests/user.json
```

Relative paths are resolved from the `.http` file location.

---

## Post-request JavaScript

Jetter supports post-request scripts with IntelliJ-style syntax:

```text
> {% client.global.set("TOKEN", response.body.token) %}
```

Supported runtime features include:
- `client.global.set("KEY", value)`
- `client.global.get("KEY")`
- `console.log(...)`
- `response.status`, `response.headers`, `response.body`

Variables set with `client.global.set(...)` can be reused in later requests via `{{KEY}}`.

---

## Jetter loop directives

Jetter adds loop directives to repeatedly execute one request until a JS condition evaluates to false.

Supported keys:
- `@jetter.while`
- `@jetter.maxIterations`
- `@jetter.sleep`
- `@jetter.onTimeout` (`fail` or `continue`)

Example:

```text
### Poll Status
GET {{URL}}/status
Authorization: Bearer {{$auth.token("auth-id")}}

#@jetter.while = response.body.status != "SUCCESS"
#@jetter.maxIterations = 30
#@jetter.sleep = 2s
#@jetter.onTimeout = fail
```

With `onTimeout = fail`, scenario execution stops after reaching `maxIterations` while the condition is still true.

---

## Logging

Use `--log` to write all request/response events in execution order to a log file.

```sh
jetter --file examples/example.http --env examples/http-client.env.json:local --log logs/requests.ndjson
```

Each entry includes:
- request method, URL, headers, body
- response status, headers, body, error (if any)
- start and finish timestamps and duration

---

## Local Testing
- See `examples/` for sample `example.http` and `examples/http-client.env.json` files.
- Use the provided `docker-compose.yml` for local Keycloak and Wiremock setup.

```sh
make local-setup
```

- Access Wiremock at `http://localhost:8080` and Keycloak at `http://localhost:8081`.
- Use the `examples/http-client.env.json` for testing authentication and API calls.

```sh
make run
```

---

## Makefile Commands

The included **Makefile** makes it easy to build, test, and run Jetter.

```sh
make <command>
```

Available commands:

| Command            | Description                                               |
|--------------------|-----------------------------------------------------------|
| `make build`       | Build the project binary (`bin/jetter`)                   |
| `make install`     | Install the binary into `~/bin`                           |
| `make run`         | Run Jetter with the example `.http` and environment files |
| `make test`        | Run all Go tests                                          |
| `make local-setup` | Start local Keycloak + Wiremock via Docker Compose        |
| `make coverage`    | Generate test coverage report                             |
| `make help`        | Print a summary of all available commands                 |

---

## 📚 Backlog & Roadmap
See [BACKLOG.md](./BACKLOG.md) for planned features and ongoing development.

---

## 🤝 Contributing
PRs and feedback are welcome! Please open issues for bugs, feature requests, or questions.

---

## License
MIT
