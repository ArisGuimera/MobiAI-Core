// Package mcp exposes MobiAI Brain as an MCP (Model Context Protocol)
// server. When wired through `mobiai brain mcp`, every brain operation
// (context, search, scan, save) becomes a tool the AI agent can invoke
// directly without the user having to run a CLI command.
//
// The server is intentionally thin: each tool handler is a wrapper over
// the existing functions in the brain package. The on-disk format,
// guards (brain-not-initialized errors) and behavior are identical to
// the CLI — only the invocation surface differs.
package mcp

import (
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/brain"
)

// ServerName is the implementation name advertised to MCP clients.
const ServerName = "mobiai-brain"

// NewServer builds an MCP server bound to the brain rooted at
// projectRoot. The root is captured once at server construction — MCP
// clients launch the server as a subprocess with a stable cwd, so the
// root doesn't change between tool calls.
//
// Returns an error if projectRoot is empty.
func NewServer(projectRoot, version string) (*mcp.Server, error) {
	if projectRoot == "" {
		return nil, fmt.Errorf("projectRoot vacío")
	}
	paths := brain.NewBrainPaths(projectRoot)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    ServerName,
		Version: version,
	}, nil)

	registerTools(server, paths)
	return server, nil
}

// registerTools wires every brain operation as an MCP tool. Each tool's
// arguments are documented inline via `jsonschema:"..."` struct tags so
// the agent sees rich descriptions in its tool catalog.
func registerTools(server *mcp.Server, paths brain.BrainPaths) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "mobile_context",
		Description: "Read the per-project MobiAI Brain context: project rules, detected " +
			"mobile stack (Android/iOS/KMP/Flutter/RN + libraries), and saved memories " +
			"(decisions, bugfixes, testing patterns, integrations, releases). " +
			"Optional filters narrow which sections and entries appear so you don't " +
			"have to dump the whole brain. Call this before proposing architecture, " +
			"integration or testing changes.",
	}, handleContext(paths))

	mcp.AddTool(server, &mcp.Tool{
		Name: "mobile_search",
		Description: "Free-text search over the brain's memory entries (case-insensitive, " +
			"matches title and body). Optional platform/status/area filters narrow " +
			"results with AND semantics. Use this to find prior decisions or known " +
			"workarounds before suggesting a similar change.",
	}, handleSearch(paths))

	mcp.AddTool(server, &mcp.Tool{
		Name: "mobile_scan",
		Description: "Re-scan the project tree and refresh the detected stack " +
			"(project type, platforms, libraries, integrations, CI/CD). The scan " +
			"never reads sensitive files (.env, Firebase configs, signing material) " +
			"— it only reports their presence. Run after large refactors or new " +
			"dependency additions.",
	}, handleScan(paths))

	mcp.AddTool(server, &mcp.Tool{
		Name: "mobile_review",
		Description: "Audit temporary debt in the brain: returns every entry with " +
			"`status: temporary` whose `review_after` date has passed. Pass " +
			"`include_no_date: true` to also surface temporary entries that lack a " +
			"review_after, listed separately under `no_date`. Use this at the start " +
			"of a session to spot workarounds that should be revisited before " +
			"proposing related changes — and to keep temporary fixes from silently " +
			"becoming permanent.",
	}, handleReview(paths))

	mcp.AddTool(server, &mcp.Tool{
		Name: "mobile_save_decision",
		Description: "Save an architecture decision to the brain (decisions.md). Use for " +
			"library/pattern choices, project-wide conventions, or tradeoffs that took " +
			"real thought. The optional `alternatives` field is the high-leverage part " +
			"— it stops future agents from re-litigating the same decision.",
	}, handleSaveDecision(paths))

	mcp.AddTool(server, &mcp.Tool{
		Name: "mobile_save_bugfix",
		Description: "Save a bugfix or workaround to the brain (bugfixes.md). Use " +
			"`status: temporary` with `review_after` for workarounds that should be " +
			"revisited; `status: active` for real fixes worth remembering for the " +
			"causal chain.",
	}, handleSaveBugfix(paths))

	mcp.AddTool(server, &mcp.Tool{
		Name: "mobile_save_testing",
		Description: "Save a reusable testing pattern to the brain (testing.md). Use " +
			"for non-obvious mobile-specific patterns: async/lifecycle timing, " +
			"platform quirks, flaky-test workarounds, fixture setups.",
	}, handleSaveTesting(paths))
}
