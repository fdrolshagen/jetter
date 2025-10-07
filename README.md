# 🚀 Jetter

[![build & test](https://github.com/fdrolshagen/jetter/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/fdrolshagen/jetter/actions/workflows/go.yml)
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
  Use the same request format you already know from IntelliJ.

- **Full IntelliJ compatibility** ([specification](https://www.jetbrains.com/help/idea/http-client-in-product-code-editor.html))  
  Leverage IntelliJ’s syntax highlighting, auto-completion, and request runner. Jetter supports:
  - 🌍 Environment variables
  - 📝 In-Place & dynamic variables
  - 📑 Multiple requests per file
  - 🔑 Authentication hooks (OAuth2)

- **Flexible execution modes**
  - Run scenarios once for quick checks
  - Simulate load with duration-based execution

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

| Flag         | Alias | Description                                                                                  |
|--------------|-------|----------------------------------------------------------------------------------------------|
| `--help`     | `-h`  | Print available command line arguments and explanation                                        |
| `--file`     | `-f`  | Path to the .http file (required)                                                            |
| `--env`      | `-e`  | Path to the environment file. Format: `-e <file>:<env-key>`                                  |
| `--duration` | `-d`  | How long should the load test run (e.g. `30s`, `1m`)                                         |
| `--version`  |   | Print version and exit                                         |

---

## Example .http File

```text
### Create User
POST {{URL}}/users
Authorization: Bearer {{$auth.token("auth-id")}}
Content-Type: application/json

{
  "username": "username",
  "email": "foobar@test.com"
}

### Get All Users
GET {{URL}}/users
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
| `{{$random.$uuid}}`  | Generates a random UUIDv4                         |
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

| Command            | Description                                                   |
|--------------------|---------------------------------------------------------------|
| `make build`       | Build the project binary (`bin/jetter`)                    |
| `make install`     | Install the binary into `~/bin`                            |
| `make run`         | Run Jetter with the example `.http` and environment files  |
| `make test`        | Run all Go tests                                           |
| `make local-setup` | Start local Keycloak + Wiremock via Docker Compose         |
| `make help`        | Print a summary of all available commands                  |

---

## 📚 Backlog & Roadmap
See [BACKLOG.md](./BACKLOG.md) for planned features and ongoing development.

---

## 🤝 Contributing
PRs and feedback are welcome! Please open issues for bugs, feature requests, or questions.

---

## License
MIT

