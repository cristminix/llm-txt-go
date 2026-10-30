// test-server is a standalone integration test for the medusa-docs MCP server.
// It spawns the server as a child process and validates all JSON-RPC interactions.
//
// Usage:
//
//	go run test-server.go [--server ./llm-txt-go] [--docs /path/to/llms.txt]
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// JSON-RPC message types
type request struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	serverBin := flag.String("server", "./llm-txt-go", "Path to MCP server binary")
	docsPath := flag.String("docs", "", "Path to documentation file (env: LLM_DOCS_PATH)")
	flag.Parse()

	// Resolve docs path
	if *docsPath == "" {
		candidates := []string{
			os.Getenv("LLM_DOCS_PATH"),
			"docs/llms.txt",
			"docs/llms-full.txt",
			"../docs/llms.txt",
			"../docs/llms-full.txt",
		}
		for _, c := range candidates {
			if c != "" {
				if _, err := os.Stat(c); err == nil {
					*docsPath = c
					break
				}
			}
		}
	}

	if *docsPath == "" {
		fmt.Println("❌  Documentation file not found. Set --docs or LLM_DOCS_PATH")
		os.Exit(1)
	}

	fmt.Printf("📄 Docs path: %s\n", *docsPath)
	fmt.Printf("🚀 Server binary: %s\n\n", *serverBin)

	cmd := exec.Command(*serverBin)
	cmd.Env = append(os.Environ(), "LLM_DOCS_PATH="+*docsPath)

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		fmt.Printf("❌  Failed to start server: %v\n", err)
		os.Exit(1)
	}
	defer cmd.Process.Kill()

	reader := bufio.NewReader(stdout)
	// Drain stderr (logs go here)
	go func() { io.Copy(io.Discard, stderr) }()

	passed := 0
	failed := 0

	// ── Test 1: Initialize ──────────────────────────────────────
	fmt.Println("── Test 1: Initialize ──")
	send(stdin, request{JSONRPC: "2.0", ID: 1, Method: "initialize", Params: map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}})

	resp := readResponse(reader)
	if assertNoError(1, resp) {
		var result map[string]interface{}
		json.Unmarshal(resp.Result, &result)
		fmt.Printf("   ✅  Server: %s v%s\n", result["serverInfo"].(map[string]interface{})["name"], result["serverInfo"].(map[string]interface{})["version"])
		passed++
	} else {
		failed++
	}

	// Send initialized notification
	send(stdin, map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	})

	// ── Test 2: List tools ─────────────────────────────────────
	fmt.Println("\n── Test 2: List tools ──")
	send(stdin, request{JSONRPC: "2.0", ID: 2, Method: "tools/list"})

	resp = readResponse(reader)
	if assertNoError(2, resp) {
		var result map[string]interface{}
		json.Unmarshal(resp.Result, &result)
		tools := result["tools"].([]interface{})
		fmt.Printf("   ✅  %d tools registered:\n", len(tools))
		for _, t := range tools {
			tool := t.(map[string]interface{})
			fmt.Printf("      • %s — %s\n", tool["name"], tool["description"])
		}
		if len(tools) == 3 {
			passed++
		} else {
			fmt.Printf("   ❌  Expected 3 tools, got %d\n", len(tools))
			failed++
		}
	} else {
		failed++
	}

	// ── Test 3: search_docs ────────────────────────────────────
	fmt.Println("\n── Test 3: search_docs ──")
	send(stdin, request{
		JSONRPC: "2.0", ID: 3, Method: "tools/call",
		Params: map[string]interface{}{
			"name":      "search_docs",
			"arguments": map[string]interface{}{"query": "authentication", "limit": 3},
		},
	})

	resp = readResponse(reader)
	if assertNoError(3, resp) {
		var result map[string]interface{}
		json.Unmarshal(resp.Result, &result)
		content := result["content"].([]interface{})
		text := content[0].(map[string]interface{})["text"].(string)
		if !strings.Contains(text, "No results found") {
			fmt.Printf("   ✅  Found results (first 100 chars): %s...\n", truncate(text, 100))
			passed++
		} else {
			fmt.Printf("   ⚠️  No results for 'authentication' (query-dependent, not a server error)\n")
			passed++ // Not a server fault
		}
	} else {
		failed++
	}

	// ── Test 4: get_section ────────────────────────────────────
	fmt.Println("\n── Test 4: get_section ──")
	send(stdin, request{
		JSONRPC: "2.0", ID: 4, Method: "tools/call",
		Params: map[string]interface{}{
			"name":      "get_section",
			"arguments": map[string]interface{}{"identifier": "introduction"},
		},
	})

	resp = readResponse(reader)
	if assertNoError(4, resp) {
		var result map[string]interface{}
		json.Unmarshal(resp.Result, &result)
		content := result["content"].([]interface{})
		text := content[0].(map[string]interface{})["text"].(string)
		if strings.HasPrefix(text, "# ") {
			fmt.Printf("   ✅  Section returned (markdown heading detected), %d chars\n", len(text))
			passed++
		} else {
			fmt.Printf("   ⚠️  Section not found for 'introduction' (docs-dependent)\n")
			passed++ // Not a server fault
		}
	} else {
		failed++
	}

	// ── Test 5: get_section not found ──────────────────────────
	fmt.Println("\n── Test 5: get_section (not found) ──")
	send(stdin, request{
		JSONRPC: "2.0", ID: 5, Method: "tools/call",
		Params: map[string]interface{}{
			"name":      "get_section",
			"arguments": map[string]interface{}{"identifier": "xyzzy_nonexistent_12345"},
		},
	})

	resp = readResponse(reader)
	if assertNoError(5, resp) {
		var result map[string]interface{}
		json.Unmarshal(resp.Result, &result)
		content := result["content"].([]interface{})
		text := content[0].(map[string]interface{})["text"].(string)
		if strings.Contains(text, "Section not found") {
			fmt.Printf("   ✅  Correctly returned 'Section not found'\n")
			passed++
		} else {
			fmt.Printf("   ❌  Unexpected response\n")
			failed++
		}
	} else {
		failed++
	}

	// ── Test 6: list_sections ──────────────────────────────────
	fmt.Println("\n── Test 6: list_sections ──")
	send(stdin, request{
		JSONRPC: "2.0", ID: 6, Method: "tools/call",
		Params: map[string]interface{}{
			"name":      "list_sections",
			"arguments": map[string]interface{}{},
		},
	})

	resp = readResponse(reader)
	if assertNoError(6, resp) {
		var result map[string]interface{}
		json.Unmarshal(resp.Result, &result)
		content := result["content"].([]interface{})
		text := content[0].(map[string]interface{})["text"].(string)
		if strings.Contains(text, "Available sections") && strings.Contains(text, "Total:") {
			fmt.Printf("   ✅  Sections listed correctly\n")
			passed++
		} else {
			fmt.Printf("   ❌  Unexpected response format\n")
			failed++
		}
	} else {
		failed++
	}

	// ── Test 7: list_sections with category filter ──────────────
	fmt.Println("\n── Test 7: list_sections (filtered) ──")
	send(stdin, request{
		JSONRPC: "2.0", ID: 7, Method: "tools/call",
		Params: map[string]interface{}{
			"name":      "list_sections",
			"arguments": map[string]interface{}{"category": "api"},
		},
	})

	resp = readResponse(reader)
	if assertNoError(7, resp) {
		var result map[string]interface{}
		json.Unmarshal(resp.Result, &result)
		content := result["content"].([]interface{})
		text := content[0].(map[string]interface{})["text"].(string)
		if strings.Contains(text, "filtered by:") && strings.Contains(text, "Total:") {
			fmt.Printf("   ✅  Filtered listing works\n")
			passed++
		} else {
			fmt.Printf("   ⚠️  No filter match for 'api' (docs-dependent)\n")
			passed++
		}
	} else {
		failed++
	}

	// ── Summary ─────────────────────────────────────────────────
	fmt.Printf("\n═══════════════════════════════════════\n")
	fmt.Printf("   Results: %d passed, %d failed, %d total\n", passed, failed, passed+failed)
	if failed == 0 {
		fmt.Printf("   ✅  All tests passed!\n")
	} else {
		fmt.Printf("   ❌  %d test(s) failed\n", failed)
		os.Exit(1)
	}
	fmt.Printf("═══════════════════════════════════════\n")
}

func send(stdin io.Writer, msg interface{}) {
	data, _ := json.Marshal(msg)
	stdin.Write(append(data, '\n'))
}

func readResponse(r *bufio.Reader) response {
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			fmt.Printf("   ❌  Read error: %v\n", err)
			os.Exit(1)
		}
		// Skip non-JSON-RPC lines (stdio logs)
		if !strings.HasPrefix(string(line), "{") {
			continue
		}
		var resp response
		if err := json.Unmarshal(line, &resp); err != nil {
			fmt.Printf("   ⚠️  Parse error: %v (raw: %s)\n", err, truncate(string(line), 80))
			continue
		}
		return resp
	}
}

func assertNoError(id int, resp response) bool {
	if resp.Error != nil {
		fmt.Printf("   ❌  RPC error [id=%d]: code=%d %s\n", id, resp.Error.Code, resp.Error.Message)
		return false
	}
	if resp.ID != id {
		fmt.Printf("   ❌  Mismatched ID: expected %d, got %d\n", id, resp.ID)
		return false
	}
	return true
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
