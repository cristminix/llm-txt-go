# llm-txt-go

A high-performance Model Context Protocol (MCP) server written in Go that serves local `llms.txt` / `llms-full.txt` documentation to AI assistants and coding agents.

This project enables LLMs and AI pair programmers to query, search, and retrieve structured sections from local documentation files efficiently over the standard `stdio` transport.

---

## Features

- **MCP Compliant**: Built on top of the Model Context Protocol (`mcp-go`) standard using `stdio`.
- **Fuzzy Search**: In-memory text matching and search scoring across titles and documentation sections.
- **Tools Provided**:
  - `search_docs`: Search documentation sections by query string with configurable match thresholds and result limits.
  - `get_section`: Retrieve the full content of a specific documentation section by title or path.
  - `list_sections`: List available sections with previews and pagination support.
- **Configurable**: Fully customizable via `config.json` and `.env` environment variables.
- **Lightweight & Fast**: Pure Go implementation with zero heavy external dependencies.

---

## Requirements

- **Go**: 1.22 or higher
- Documentation file in `llms-full.txt` or `llms.txt` format

---

## Installation & Build

Clone the repository:

```bash
git clone https://github.com/cristminix/llm-txt-go.git
cd llm-txt-go
```

Download dependencies:

```bash
go mod download
```

Build the binary:

```bash
go build -o bin/llm-txt-go main.go
```

---

## Configuration

### 1. `config.json`

Create or customize `config.json` in the root directory:

```json
{
  "searchDefaults": {
    "maxResults": 5,
    "threshold": 0.4,
    "minMatchCharLength": 3
  },
  "listDefaults": {
    "maxSections": 50
  },
  "server": {
    "name": "llm-txt-docs",
    "version": "1.0.0"
  },
  "documentation": {
    "previewLength": 500,
    "fallbackPaths": [
      "docs/llms-full.txt"
    ]
  }
}
```

### 2. Environment Variables (`.env`)

Copy the example environment configuration:

```bash
cp .env-example .env
```

Contents of `.env-example`:
```bash
DOCUMENTATION_NAME="Medusa Docs"
LLM_DOCS_PATH=docs/llms.txt
```

Available environment variables:
- `DOCUMENTATION_NAME`: Display name for the documentation server (default: `Medusa Docs`).
- `LLM_DOCS_PATH`: Path to your documentation file (e.g., `docs/llms-full.txt` or `docs/llms.txt`).
- `CONFIG_PATH`: Optional custom path to `config.json` (defaults to `./config.json`).

---

## Usage

### Running Locally

Run directly:

```bash
go run main.go
```

Or using the compiled binary:

```bash
./bin/llm-txt-go
```

### Adding to MCP Clients (Claude Desktop, Antigravity, Cursor, etc.)

Add the server to your MCP client configuration (e.g. `claude_desktop_config.json` or your agent's MCP settings):

```json
{
  "mcpServers": {
    "llm-docs": {
      "command": "/absolute/path/to/llm-txt-go/bin/llm-txt-go",
      "args": [],
      "env": {
        "DOCUMENTATION_NAME": "My Project Docs"
      }
    }
  }
}
```

---

## Available MCP Tools

| Tool | Description | Arguments |
| :--- | :--- | :--- |
| `search_docs` | Search documentation sections using keyword / fuzzy match | `query` (string, required), `maxResults` (number, optional) |
| `get_section` | Retrieve complete content of a section | `title` (string, required) |
| `list_sections` | List sections with brief previews | `limit` (number, optional), `offset` (number, optional) |

---

## Running Tests

Run the unit test suite:

```bash
go test ./...
```

---

## License

MIT License. Feel free to use and adapt this project for your own documentation needs.
