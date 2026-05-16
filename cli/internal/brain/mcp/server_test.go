package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/brain"
)

// setupClient spins up an in-memory MCP client connected to a server
// rooted at a freshly initialized brain. The brain lives in a TempDir
// so tests are hermetic.
func setupClient(t *testing.T) (*sdkmcp.ClientSession, brain.BrainPaths, func()) {
	t.Helper()
	ctx := context.Background()
	tmp := t.TempDir()
	paths := brain.NewBrainPaths(tmp)
	if err := paths.EnsureDirs(); err != nil {
		t.Fatal(err)
	}
	cfg := brain.DefaultConfig(tmp)
	if err := cfg.Save(paths); err != nil {
		t.Fatal(err)
	}
	if _, err := brain.EnsureMemoryFiles(paths); err != nil {
		t.Fatal(err)
	}

	server, err := NewServer(tmp, "test")
	if err != nil {
		t.Fatal(err)
	}
	clientTransport, serverTransport := sdkmcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test-client"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		_ = clientSession.Close()
		_ = serverSession.Wait()
	}
	return clientSession, paths, cleanup
}

func TestServer_ListsAllTools(t *testing.T) {
	client, _, cleanup := setupClient(t)
	defer cleanup()

	res, err := client.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"mobile_context":       false,
		"mobile_search":        false,
		"mobile_scan":          false,
		"mobile_review":        false,
		"mobile_promote":       false,
		"mobile_bump":          false,
		"mobile_save_decision": false,
		"mobile_save_bugfix":   false,
		"mobile_save_testing":  false,
	}
	for _, tool := range res.Tools {
		if _, ok := want[tool.Name]; ok {
			want[tool.Name] = true
		}
	}
	for name, present := range want {
		if !present {
			t.Errorf("tool %q missing from ListTools", name)
		}
	}
}

func TestTool_Context_ReturnsMarkdown(t *testing.T) {
	client, _, cleanup := setupClient(t)
	defer cleanup()

	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "mobile_context",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	text := firstTextContent(res)
	for _, want := range []string{
		"# MobiAI Brain Context",
		"## Project Rules",
		"## Architecture Decisions",
		"## Warnings",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in context output:\n%s", want, text)
		}
	}
}

func TestTool_Context_AppliesFilters(t *testing.T) {
	client, paths, cleanup := setupClient(t)
	defer cleanup()

	seedDecision(t, paths, "Use Koin", "shared")
	seedDecision(t, paths, "Use ViewModel", "android")

	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name: "mobile_context",
		Arguments: map[string]any{
			"sections": []string{"decisions"},
			"platform": "shared",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	text := firstTextContent(res)
	if !strings.Contains(text, "Use Koin") {
		t.Errorf("shared decision should pass filter:\n%s", text)
	}
	if strings.Contains(text, "Use ViewModel") {
		t.Errorf("android decision should be filtered out:\n%s", text)
	}
	if strings.Contains(text, "## Detected Stack") {
		t.Errorf("stack section should be excluded by --section decisions:\n%s", text)
	}
}

func TestTool_Search_ReturnsStructuredHits(t *testing.T) {
	client, paths, cleanup := setupClient(t)
	defer cleanup()

	seedDecision(t, paths, "Use Koin for DI", "shared")
	seedBugfix(t, paths, "Compose recomposition loop", "android")

	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "mobile_search",
		Arguments: map[string]any{"query": "koin"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// The structured output is on res.StructuredContent — round-trip
	// through JSON to verify the schema is what we declared.
	hits := decodeSearchHits(t, res)
	if len(hits) != 1 || hits[0].Title != "Use Koin for DI" {
		t.Errorf("expected 1 koin hit, got %d: %+v", len(hits), hits)
	}
	if hits[0].Section != "decisions" {
		t.Errorf("section = %q, want decisions", hits[0].Section)
	}
	if hits[0].Platform != "shared" {
		t.Errorf("platform = %q, want shared", hits[0].Platform)
	}
}

func TestTool_Search_QueryRequired(t *testing.T) {
	client, _, cleanup := setupClient(t)
	defer cleanup()

	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "mobile_search",
		Arguments: map[string]any{},
	})
	// The SDK surfaces tool errors either as a transport error or via
	// res.IsError. Accept either — we just want to confirm the empty
	// query is rejected somewhere.
	if err == nil && (res == nil || !res.IsError) {
		t.Errorf("empty query should fail, got success: %+v", res)
	}
}

func TestTool_SaveDecision_WritesEntry(t *testing.T) {
	client, paths, cleanup := setupClient(t)
	defer cleanup()

	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name: "mobile_save_decision",
		Arguments: map[string]any{
			"title":    "Use Koin as DI",
			"platform": "shared",
			"area":     "dependency_injection",
			"body":     "### Decision\nUse Koin.",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("save failed: %v", firstTextContent(res))
	}
	got, err := os.ReadFile(filepath.Join(paths.MemoriesDir, "decisions.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"## Use Koin as DI",
		"- type: architecture_decision",
		"- platform: shared",
	} {
		if !strings.Contains(string(got), want) {
			t.Errorf("missing %q in decisions.md:\n%s", want, got)
		}
	}
}

func TestTool_SaveBugfix_TemporaryStatusMapsCorrectly(t *testing.T) {
	client, paths, cleanup := setupClient(t)
	defer cleanup()

	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name: "mobile_save_bugfix",
		Arguments: map[string]any{
			"title":        "FirebaseAuth iOS rename",
			"platform":     "ios",
			"status":       "temporary",
			"review_after": "2026-12-01",
			"body":         "### Problem\n...",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("save failed: %v", firstTextContent(res))
	}
	got, _ := os.ReadFile(filepath.Join(paths.MemoriesDir, "bugfixes.md"))
	if !strings.Contains(string(got), "- type: platform_workaround") {
		t.Errorf("temporary bugfix should map to platform_workaround:\n%s", got)
	}
	if !strings.Contains(string(got), "- review_after: 2026-12-01") {
		t.Errorf("review_after missing:\n%s", got)
	}
}

func TestTool_Review_ReturnsOverdueOnly(t *testing.T) {
	client, paths, cleanup := setupClient(t)
	defer cleanup()

	// Plant three temporary entries: one overdue, one future, one no-date.
	mustSaveBugfix(t, paths, "Overdue workaround", "ios", "2024-01-01")
	mustSaveBugfix(t, paths, "Future review", "android", "2099-01-01")
	mustSaveBugfix(t, paths, "No date temp", "shared", "")

	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "mobile_review",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("review failed: %v", firstTextContent(res))
	}
	out := decodeReviewResult(t, res)
	if len(out.Overdue) != 1 {
		t.Fatalf("expected 1 overdue, got %d: %+v", len(out.Overdue), out.Overdue)
	}
	if out.Overdue[0].Title != "Overdue workaround" {
		t.Errorf("wrong overdue entry: %q", out.Overdue[0].Title)
	}
	if out.Overdue[0].DaysOverdue <= 0 {
		t.Errorf("DaysOverdue should be positive, got %d", out.Overdue[0].DaysOverdue)
	}
	if len(out.NoDate) != 0 {
		t.Errorf("no_date should be empty without include_no_date; got %+v", out.NoDate)
	}
}

func TestTool_Review_IncludeNoDate(t *testing.T) {
	client, paths, cleanup := setupClient(t)
	defer cleanup()

	mustSaveBugfix(t, paths, "Overdue workaround", "ios", "2024-01-01")
	mustSaveBugfix(t, paths, "No date temp", "shared", "")

	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "mobile_review",
		Arguments: map[string]any{"include_no_date": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("review failed: %v", firstTextContent(res))
	}
	out := decodeReviewResult(t, res)
	if len(out.Overdue) != 1 {
		t.Errorf("expected 1 overdue, got %d", len(out.Overdue))
	}
	if len(out.NoDate) != 1 || out.NoDate[0].Title != "No date temp" {
		t.Errorf("expected 1 no_date entry; got %+v", out.NoDate)
	}
}

func TestTool_Review_EmptyBrainSucceeds(t *testing.T) {
	client, _, cleanup := setupClient(t)
	defer cleanup()

	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "mobile_review",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("review on empty brain should succeed: %v", firstTextContent(res))
	}
	out := decodeReviewResult(t, res)
	if len(out.Overdue) != 0 || len(out.NoDate) != 0 {
		t.Errorf("empty brain should produce no items; got overdue=%d no_date=%d",
			len(out.Overdue), len(out.NoDate))
	}
}

func TestTool_Promote_ChangesStatusAndType(t *testing.T) {
	client, paths, cleanup := setupClient(t)
	defer cleanup()

	id, err := brain.AppendEntry(paths, &brain.SaveEntry{
		Type:        brain.SaveTypeBugfix,
		Title:       "WA",
		Status:      "temporary",
		ReviewAfter: "2026-06-01",
		Body:        "body",
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name: "mobile_promote",
		Arguments: map[string]any{
			"id":                 id,
			"status":             "active",
			"clear_review_after": true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("promote failed: %v", firstTextContent(res))
	}
	out := decodeUpdateResult(t, res)
	if out.PrevStatus != "temporary" || out.NewStatus != "active" {
		t.Errorf("status transition: prev=%q new=%q", out.PrevStatus, out.NewStatus)
	}
	if out.NewReviewAfter != "" {
		t.Errorf("review_after should be cleared; got %q", out.NewReviewAfter)
	}
	// Confirm on disk that the bugfix type was re-derived.
	data, _ := os.ReadFile(filepath.Join(paths.MemoriesDir, "bugfixes.md"))
	if !strings.Contains(string(data), "- type: bug_fix") {
		t.Errorf("type should flip to bug_fix:\n%s", data)
	}
}

func TestTool_Bump_ExtendsReviewAfter(t *testing.T) {
	client, paths, cleanup := setupClient(t)
	defer cleanup()

	id, err := brain.AppendEntry(paths, &brain.SaveEntry{
		Type:        brain.SaveTypeBugfix,
		Title:       "WA",
		Status:      "temporary",
		ReviewAfter: "2026-03-01",
		Body:        "body",
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name: "mobile_bump",
		Arguments: map[string]any{
			"id":           id,
			"review_after": "2027-01-01",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("bump failed: %v", firstTextContent(res))
	}
	out := decodeUpdateResult(t, res)
	if out.PrevReviewAfter != "2026-03-01" || out.NewReviewAfter != "2027-01-01" {
		t.Errorf("review_after transition: prev=%q new=%q", out.PrevReviewAfter, out.NewReviewAfter)
	}
	if out.NewStatus != out.PrevStatus {
		t.Errorf("bump must not change status; prev=%q new=%q", out.PrevStatus, out.NewStatus)
	}
}

func TestTool_Promote_UnknownIDFails(t *testing.T) {
	client, _, cleanup := setupClient(t)
	defer cleanup()

	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name: "mobile_promote",
		Arguments: map[string]any{
			"id":     "ghost-id",
			"status": "active",
		},
	})
	if err == nil && (res == nil || !res.IsError) {
		t.Errorf("unknown id should fail; got: %+v", res)
	}
}

func TestTool_Scan_ReturnsStructuredSummary(t *testing.T) {
	client, paths, cleanup := setupClient(t)
	defer cleanup()
	// Plant a KMP project under the same root so scan finds something.
	mustWriteFile(t, filepath.Join(paths.Root, "settings.gradle.kts"), "")
	mustWriteFile(t, filepath.Join(paths.Root, "composeApp/build.gradle.kts"),
		`plugins { kotlin("multiplatform") }
dependencies { implementation("io.insert-koin:koin-core:3.5.0") }
`)
	if err := os.MkdirAll(filepath.Join(paths.Root, "composeApp", "src", "commonMain"), 0o755); err != nil {
		t.Fatal(err)
	}

	res, err := client.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "mobile_scan",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("scan failed: %v", firstTextContent(res))
	}
	scan := decodeScanResult(t, res)
	if scan.ProjectType != "kmp" {
		t.Errorf("project_type = %q, want kmp", scan.ProjectType)
	}
	if !sliceContains(scan.DI, "koin") {
		t.Errorf("di = %v, want koin", scan.DI)
	}
	// Verify the scan was also persisted so subsequent context calls
	// see the fresh data.
	if _, err := os.Stat(paths.ScanFile); err != nil {
		t.Errorf("scan.json not persisted: %v", err)
	}
}

func TestTool_FailsCleanlyWhenBrainMissing(t *testing.T) {
	// Spin up a server pointed at a directory without a brain.
	ctx := context.Background()
	tmp := t.TempDir()
	server, err := NewServer(tmp, "test")
	if err != nil {
		t.Fatal(err)
	}
	clientTransport, serverTransport := sdkmcp.NewInMemoryTransports()
	ss, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Wait()
	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test-client"}, nil)
	cs, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()

	res, err := cs.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "mobile_context",
		Arguments: map[string]any{},
	})
	if err == nil && (res == nil || !res.IsError) {
		t.Errorf("expected error when brain not initialized; got: %+v", res)
	}
	// The error message should hint at `mobiai brain init` so the user
	// knows the recovery path.
	if err != nil {
		if !strings.Contains(err.Error(), "brain") {
			t.Errorf("error should mention brain: %v", err)
		}
	}
}

// --- helpers ---

func firstTextContent(res *sdkmcp.CallToolResult) string {
	if res == nil {
		return ""
	}
	for _, c := range res.Content {
		if tc, ok := c.(*sdkmcp.TextContent); ok {
			return tc.Text
		}
	}
	return ""
}

func decodeSearchHits(t *testing.T, res *sdkmcp.CallToolResult) []SearchHit {
	t.Helper()
	if res == nil || res.StructuredContent == nil {
		t.Fatal("expected StructuredContent")
	}
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var out SearchResult
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decode structured content: %v\nraw: %s", err, raw)
	}
	return out.Hits
}

func decodeScanResult(t *testing.T, res *sdkmcp.CallToolResult) ScanResult {
	t.Helper()
	if res == nil || res.StructuredContent == nil {
		t.Fatal("expected StructuredContent")
	}
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var out ScanResult
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decode structured content: %v\nraw: %s", err, raw)
	}
	return out
}

func seedDecision(t *testing.T, p brain.BrainPaths, title, platform string) {
	t.Helper()
	if _, err := brain.AppendEntry(p, &brain.SaveEntry{
		Type:     brain.SaveTypeDecision,
		Title:    title,
		Platform: platform,
		Body:     "body of " + title,
	}); err != nil {
		t.Fatal(err)
	}
}

// mustSaveBugfix appends a temporary bugfix entry. Empty reviewAfter
// means "save without a review_after meta field" — used by the review
// tests to plant a no-date entry.
func mustSaveBugfix(t *testing.T, p brain.BrainPaths, title, platform, reviewAfter string) {
	t.Helper()
	if _, err := brain.AppendEntry(p, &brain.SaveEntry{
		Type:        brain.SaveTypeBugfix,
		Title:       title,
		Platform:    platform,
		Status:      "temporary",
		ReviewAfter: reviewAfter,
		Body:        "body of " + title,
	}); err != nil {
		t.Fatal(err)
	}
}

func decodeUpdateResult(t *testing.T, res *sdkmcp.CallToolResult) UpdateResult {
	t.Helper()
	if res == nil || res.StructuredContent == nil {
		t.Fatal("expected StructuredContent")
	}
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var out UpdateResult
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decode structured content: %v\nraw: %s", err, raw)
	}
	return out
}

func decodeReviewResult(t *testing.T, res *sdkmcp.CallToolResult) ReviewResult {
	t.Helper()
	if res == nil || res.StructuredContent == nil {
		t.Fatal("expected StructuredContent")
	}
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var out ReviewResult
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decode structured content: %v\nraw: %s", err, raw)
	}
	return out
}

func seedBugfix(t *testing.T, p brain.BrainPaths, title, platform string) {
	t.Helper()
	if _, err := brain.AppendEntry(p, &brain.SaveEntry{
		Type:     brain.SaveTypeBugfix,
		Title:    title,
		Platform: platform,
		Body:     "body",
	}); err != nil {
		t.Fatal(err)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func sliceContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
