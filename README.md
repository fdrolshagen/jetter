# 🚀 Jetter

**Jetter** is a light-weight load testing and API scenario runner for HTTP services. It uses a subset of the IntelliJ `.http` file syntax and supports environment variables, authentication, and more.

> **:exclamation: This project is under construction and has not yet published a working version.**

---

## ✨ Features
- Parse and run `.http` scenario files
- Intellij compatible features and syntax
  - Environment variables
  - Inline variables
  - Magic variables
  - Multiple requests per file
  - Authentication hooks (OAuth2)
- Duration-based load testing or just once execution

---

## ⚡ Quick Start
```sh
jetter --file examples/example.http --env examples/http-client.env.json:local
```

---

## 🛠️ Command Line Flags

| Flag                | Alias | Description                                                                                  |
|---------------------|-------|----------------------------------------------------------------------------------------------|
| `--help`            | `-h`  | Print available command line arguments and explanation                                        |
| `--file`            | `-f`  | Path to the .http file (required)                                                            |
| `--env`             | `-e`  | Path to the environment file. Format: `-e <file>:<env-key>`                                  |
| `--duration`        | `-d`  | How long should the load test run (e.g. `30s`, `1m`)                                         |

---

## 📄 Example .http File

```http
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

## 🌱 Example Environment File

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

## 📝 Inline Variables

You can define variables directly at the top of your `.http` file using the `@` syntax:

```http
@ID = 123
@TOKEN = abc

### Get User
GET http://localhost:8081/users/{{ID}}
Authorization: Bearer {{TOKEN}}
```

- Inline variables are available for substitution in URLs, headers, and bodies.
- Environment variables (from `-e`) are also available, but **inline variables take precedence** if keys overlap.

---

## 💡 Magic Variables

Described below are some built-in magic variables you can use in your `.http` files.

| Variable                     | Description                                       |
|------------------------------|---------------------------------------------------|
| `{{$random.$uuid}}`  | Generates a random UUIDv4                         |
| `{{$random.hexadecimal(n)}}` | Generates a random hexadecimal string of length n |

---

## Authenticantion Hooks
Jetter supports Oauth2 authentication out of the box. You can define multiple auth configurations in your environment file and reference them in your `.http` file using the `{{$auth.token("auth-id")}}` magic variable. Supported Grant Types: `Client Credentials` and `Password`.

## 🧪 Local Testing
- See `examples/` for sample `.http` and environment files.
- Use the provided `docker-compose.yml` for local Keycloak and Wiremock setup.

```sh
make local-setup
```

- Access Wiremock at `http://localhost:8081` and Keycloak at `http://localhost:8080`.
- Use the `examples/http-client.env.json` for testing authentication and API calls.

```sh
make run
```

---

## 📚 Backlog & Roadmap
See [BACKLOG.md](./BACKLOG.md) for planned features and ongoing development.

---

## 🤝 Contributing
PRs and feedback are welcome! Please open issues for bugs, feature requests, or questions.

---



---

## License
MIT

