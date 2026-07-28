package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/brain"
)

// ContextArgs are the inputs for mobile_context. Every field is
// optional — the zero value renders the full document. Field names and
// semantics mirror the corresponding CLI flags.
type ContextArgs struct {
	Sections []string `json:"sections,omitempty" jsonschema:"limit which sections to render (stack, rules, decisions, bugfixes, testing, integrations, releases, warnings). Empty = render all."`
	Platform string   `json:"platform,omitempty" jsonschema:"filter memory entries by platform (android|ios|shared|kmp|flutter|react-native). Exact, case-insensitive match."`
	Status   string   `json:"status,omitempty" jsonschema:"filter memory entries by status (active|temporary|deprecated). Exact match."`
	Area     string   `json:"area,omitempty" jsonschema:"filter memory entries whose area contains this substring."`
}

// ContextResult returns the rendered Markdown so MCP clients can show
// it verbatim. Markdown (not JSON) is intentional — the document IS the
// context the agent needs to read.
type ContextResult struct {
	Markdown string `json:"markdown"`
}

func handleContext(paths brain.BrainPaths) func(context.Context, *mcp.CallToolRequest, ContextArgs) (*mcp.CallToolResult, ContextResult, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, args ContextArgs) (*mcp.CallToolResult, ContextResult, error) {
		if !paths.Exists() {
			return nil, ContextResult{}, brainNotInitialized(paths)
		}
		cfg, err := brain.LoadConfig(paths)
		if err != nil {
			return nil, ContextResult{}, err
		}
		scan, err := brain.LoadScan(paths)
		if err != nil {
			return nil, ContextResult{}, err
		}
		md := brain.BuildContextWith(cfg, scan, paths, brain.ContextOptions{
			Sections: args.Sections,
			Filter: brain.EntryFilter{
				Platform: args.Platform,
				Status:   args.Status,
				Area:     args.Area,
			},
		})
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: md}},
		}, ContextResult{Markdown: md}, nil
	}
}

// SearchArgs are the inputs for mobile_search. Query is required;
// filters are optional and AND-composed with the query.
type SearchArgs struct {
	Query    string `json:"query" jsonschema:"free-text to match against entry title and body (case-insensitive substring)"`
	Platform string `json:"platform,omitempty" jsonschema:"limit to entries with this platform"`
	Status   string `json:"status,omitempty" jsonschema:"limit to entries with this status"`
	Area     string `json:"area,omitempty" jsonschema:"limit to entries whose area contains this substring"`
}

// SearchHit is the JSON-friendly view of brain.SearchHit. We flatten
// the metadata onto the hit so the agent doesn't need to know the
// internal entry structure.
type SearchHit struct {
	Section  string `json:"section"`
	Title    string `json:"title"`
	ID       string `json:"id,omitempty"`
	Status   string `json:"status,omitempty"`
	Platform string `json:"platform,omitempty"`
	Area     string `json:"area,omitempty"`
	Snippet  string `json:"snippet,omitempty"`
}

// SearchResult is the structured output of mobile_search. Returning
// JSON (not text) lets agents sort, cite and follow up on specific
// hits without re-parsing prose.
type SearchResult struct {
	Hits []SearchHit `json:"hits"`
}

func handleSearch(paths brain.BrainPaths) func(context.Context, *mcp.CallToolRequest, SearchArgs) (*mcp.CallToolResult, SearchResult, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, args SearchArgs) (*mcp.CallToolResult, SearchResult, error) {
		if !paths.Exists() {
			return nil, SearchResult{}, brainNotInitialized(paths)
		}
		if strings.TrimSpace(args.Query) == "" {
			return nil, SearchResult{}, fmt.Errorf("query is required")
		}
		raw, err := brain.Search(paths, args.Query, brain.EntryFilter{
			Platform: args.Platform,
			Status:   args.Status,
			Area:     args.Area,
		})
		if err != nil {
			return nil, SearchResult{}, err
		}
		out := SearchResult{Hits: make([]SearchHit, 0, len(raw))}
		for _, h := range raw {
			out.Hits = append(out.Hits, SearchHit{
				Section:  h.Section,
				Title:    h.Entry.Title,
				ID:       h.Entry.Get("id"),
				Status:   h.Entry.Get("status"),
				Platform: h.Entry.Get("platform"),
				Area:     h.Entry.Get("area"),
				Snippet:  h.Snippet,
			})
		}
		// Also include a short text summary in Content so MCP clients
		// that show tool output verbatim render something useful.
		summary := formatSearchSummary(args.Query, out.Hits)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: summary}},
		}, out, nil
	}
}

func formatSearchSummary(query string, hits []SearchHit) string {
	if len(hits) == 0 {
		return fmt.Sprintf("No results for %q.", query)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d result(s) for %q:\n", len(hits), query)
	for _, h := range hits {
		var meta []string
		if h.Status != "" {
			meta = append(meta, h.Status)
		}
		if h.Platform != "" {
			meta = append(meta, h.Platform)
		}
		suffix := ""
		if len(meta) > 0 {
			suffix = " (" + strings.Join(meta, ", ") + ")"
		}
		fmt.Fprintf(&b, "[%s] %s%s — %s\n", h.Section, h.Title, suffix, h.Snippet)
	}
	return b.String()
}

// ScanArgs is intentionally empty — scan takes the project as it is.
type ScanArgs struct{}

// ScanResult is the JSON-friendly slice of fields agents care about
// from a scan. We don't dump the full scan.json (it has bookkeeping
// fields like created_at that aren't actionable for an agent).
type ScanResult struct {
	ProjectType   string            `json:"project_type"`
	Platforms     []string          `json:"platforms"`
	BuildSystems  []string          `json:"build_systems,omitempty"`
	UI            []string          `json:"ui,omitempty"`
	DI            []string          `json:"di,omitempty"`
	Network       []string          `json:"network,omitempty"`
	Persistence   []string          `json:"persistence,omitempty"`
	Serialization []string          `json:"serialization,omitempty"`
	Testing       []string          `json:"testing,omitempty"`
	Integrations  []string          `json:"integrations,omitempty"`
	CICD          []string          `json:"ci_cd,omitempty"`
	DetectedFiles map[string]string `json:"detected_files,omitempty"`
	Warnings      []string          `json:"warnings,omitempty"`
}

func handleScan(paths brain.BrainPaths) func(context.Context, *mcp.CallToolRequest, ScanArgs) (*mcp.CallToolResult, ScanResult, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, _ ScanArgs) (*mcp.CallToolResult, ScanResult, error) {
		if !paths.Exists() {
			return nil, ScanResult{}, brainNotInitialized(paths)
		}
		scan, err := brain.ScanProject(paths.Root)
		if err != nil {
			return nil, ScanResult{}, err
		}
		// Persist so subsequent mobile_context reflects the fresh scan
		// without the agent needing to call anything else.
		if err := scan.Save(paths); err != nil {
			return nil, ScanResult{}, err
		}
		out := ScanResult{
			ProjectType:   scan.ProjectType,
			Platforms:     scan.Platforms,
			BuildSystems:  scan.BuildSystems,
			UI:            scan.UI,
			DI:            scan.DI,
			Network:       scan.Network,
			Persistence:   scan.Persistence,
			Serialization: scan.Serialization,
			Testing:       scan.Testing,
			Integrations:  scan.Integrations,
			CICD:          scan.CICD,
			DetectedFiles: scan.DetectedFiles,
			Warnings:      scan.Warnings,
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: formatScanSummary(scan)}},
		}, out, nil
	}
}

func formatScanSummary(s *brain.Scan) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Type: %s\n", s.ProjectType)
	if len(s.Platforms) > 0 {
		fmt.Fprintf(&b, "Platforms: %s\n", strings.Join(s.Platforms, ", "))
	}
	for _, row := range []struct {
		label string
		items []string
	}{
		{"UI", s.UI},
		{"DI", s.DI},
		{"Network", s.Network},
		{"Integrations", s.Integrations},
	} {
		if len(row.items) > 0 {
			fmt.Fprintf(&b, "%s: %s\n", row.label, strings.Join(row.items, ", "))
		}
	}
	if len(s.Warnings) > 0 {
		fmt.Fprintf(&b, "Warnings: %d\n", len(s.Warnings))
	}
	return b.String()
}

// ReviewArgs configures mobile_review. Both fields are optional; the
// zero value mirrors the CLI default (only overdue dated entries).
type ReviewArgs struct {
	IncludeNoDate bool `json:"include_no_date,omitempty" jsonschema:"also list temporary entries that have no review_after date. They land under no_date in the result. Defaults to false."`
}

// ReviewItem is the JSON-friendly view of brain.ReviewItem. Same
// flattening philosophy as SearchHit — agents don't need to know the
// raw entry structure.
type ReviewItem struct {
	Section     string `json:"section"`
	Title       string `json:"title"`
	ID          string `json:"id,omitempty"`
	Platform    string `json:"platform,omitempty"`
	Area        string `json:"area,omitempty"`
	ReviewAfter string `json:"review_after,omitempty"`
	DaysOverdue int    `json:"days_overdue"`
}

// ReviewResult separates overdue from no-date entries so agents can
// reason about them independently — "we have debt that's past due" is a
// different signal from "we have debt with no review date set".
type ReviewResult struct {
	Overdue []ReviewItem `json:"overdue"`
	NoDate  []ReviewItem `json:"no_date,omitempty"`
}

func handleReview(paths brain.BrainPaths) func(context.Context, *mcp.CallToolRequest, ReviewArgs) (*mcp.CallToolResult, ReviewResult, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, args ReviewArgs) (*mcp.CallToolResult, ReviewResult, error) {
		if !paths.Exists() {
			return nil, ReviewResult{}, brainNotInitialized(paths)
		}
		raw, err := brain.Review(paths, brain.ReviewOptions{
			IncludeNoDate: args.IncludeNoDate,
		})
		if err != nil {
			return nil, ReviewResult{}, err
		}
		out := ReviewResult{
			Overdue: make([]ReviewItem, 0, len(raw)),
		}
		for _, it := range raw {
			item := ReviewItem{
				Section:     it.Section,
				Title:       it.Entry.Title,
				ID:          it.Entry.Get("id"),
				Platform:    it.Entry.Get("platform"),
				Area:        it.Entry.Get("area"),
				ReviewAfter: it.ReviewAfter,
				DaysOverdue: it.DaysOverdue,
			}
			if it.HasDate {
				out.Overdue = append(out.Overdue, item)
			} else {
				out.NoDate = append(out.NoDate, item)
			}
		}
		// Short text summary for MCP clients that show tool output
		// verbatim. Structured data lives on the second return value.
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: formatReviewSummary(out)}},
		}, out, nil
	}
}

func formatReviewSummary(r ReviewResult) string {
	if len(r.Overdue) == 0 && len(r.NoDate) == 0 {
		return "✓ No expired temporary entries."
	}
	var b strings.Builder
	if len(r.Overdue) > 0 {
		fmt.Fprintf(&b, "⚠ %d expired entry(ies):\n", len(r.Overdue))
		for _, it := range r.Overdue {
			fmt.Fprintf(&b, "  [%s] %s — review_after: %s (%d days)\n",
				it.Section, it.Title, it.ReviewAfter, it.DaysOverdue)
		}
	}
	if len(r.NoDate) > 0 {
		fmt.Fprintf(&b, "\n%d without review_after:\n", len(r.NoDate))
		for _, it := range r.NoDate {
			fmt.Fprintf(&b, "  [%s] %s\n", it.Section, it.Title)
		}
	}
	return b.String()
}

// PromoteArgs configures mobile_promote. Status is required (the whole
// point is changing status); review_after and clear_review_after are
// optional follow-up fields for the common case where promoting and
// extending/clearing the review window happens in one shot.
type PromoteArgs struct {
	ID               string `json:"id" jsonschema:"the id of the entry to update (find via mobile_search or mobile_review)"`
	Status           string `json:"status" jsonschema:"new status: active | temporary | deprecated"`
	ReviewAfter      string `json:"review_after,omitempty" jsonschema:"also set review_after to this YYYY-MM-DD date (mutually exclusive with clear_review_after)"`
	ClearReviewAfter bool   `json:"clear_review_after,omitempty" jsonschema:"remove review_after entirely (mutually exclusive with review_after)"`
}

// BumpArgs configures mobile_bump. Both fields are required — bumping
// is just extending review_after on an existing entry without touching
// status. For status changes, use mobile_promote.
type BumpArgs struct {
	ID          string `json:"id" jsonschema:"the id of the entry whose review_after to extend"`
	ReviewAfter string `json:"review_after" jsonschema:"new review_after as YYYY-MM-DD"`
}

// UpdateResult is the JSON-friendly view of brain.UpdateResult. Same
// flattening philosophy as SearchHit: agents see what changed without
// reading raw entry structures.
type UpdateResult struct {
	Section         string `json:"section"`
	File            string `json:"file"`
	Title           string `json:"title"`
	PrevStatus      string `json:"prev_status,omitempty"`
	NewStatus       string `json:"new_status,omitempty"`
	PrevReviewAfter string `json:"prev_review_after,omitempty"`
	NewReviewAfter  string `json:"new_review_after,omitempty"`
}

func handlePromote(paths brain.BrainPaths) func(context.Context, *mcp.CallToolRequest, PromoteArgs) (*mcp.CallToolResult, UpdateResult, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, args PromoteArgs) (*mcp.CallToolResult, UpdateResult, error) {
		res, err := brain.UpdateEntry(paths, args.ID, brain.UpdateOptions{
			Status:           args.Status,
			ReviewAfter:      args.ReviewAfter,
			ClearReviewAfter: args.ClearReviewAfter,
		})
		if err != nil {
			return nil, UpdateResult{}, err
		}
		out := updateResultToJSON(res)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: formatUpdateSummary(res)}},
		}, out, nil
	}
}

func handleBump(paths brain.BrainPaths) func(context.Context, *mcp.CallToolRequest, BumpArgs) (*mcp.CallToolResult, UpdateResult, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, args BumpArgs) (*mcp.CallToolResult, UpdateResult, error) {
		res, err := brain.UpdateEntry(paths, args.ID, brain.UpdateOptions{
			ReviewAfter: args.ReviewAfter,
		})
		if err != nil {
			return nil, UpdateResult{}, err
		}
		out := updateResultToJSON(res)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: formatUpdateSummary(res)}},
		}, out, nil
	}
}

func updateResultToJSON(r *brain.UpdateResult) UpdateResult {
	return UpdateResult{
		Section:         r.Section,
		File:            r.File,
		Title:           r.Title,
		PrevStatus:      r.PrevStatus,
		NewStatus:       r.NewStatus,
		PrevReviewAfter: r.PrevReviewAfter,
		NewReviewAfter:  r.NewReviewAfter,
	}
}

func formatUpdateSummary(r *brain.UpdateResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "✓ %s updated (%s)", r.Title, r.File)
	if r.PrevStatus != r.NewStatus {
		fmt.Fprintf(&b, "\n  status: %s → %s", r.PrevStatus, r.NewStatus)
	}
	if r.PrevReviewAfter != r.NewReviewAfter {
		from, to := r.PrevReviewAfter, r.NewReviewAfter
		if from == "" {
			from = "(none)"
		}
		if to == "" {
			to = "(none)"
		}
		fmt.Fprintf(&b, "\n  review_after: %s → %s", from, to)
	}
	return b.String()
}

// SaveArgs is the shared shape for the three save_* tools. Unused
// fields per tool (e.g. ReviewAfter on decision saves) are simply
// ignored downstream — the per-tool description tells the agent which
// fields are meaningful.
type SaveArgs struct {
	Title       string   `json:"title" jsonschema:"short descriptive title (becomes the H2 heading of the entry)"`
	Platform    string   `json:"platform,omitempty" jsonschema:"android|ios|shared|kmp|flutter|react-native"`
	Area        string   `json:"area,omitempty" jsonschema:"free-form, e.g. firebase_auth | dependency_injection | datastore"`
	Status      string   `json:"status,omitempty" jsonschema:"active (default) | temporary | deprecated"`
	ReviewAfter string   `json:"review_after,omitempty" jsonschema:"YYYY-MM-DD — mostly meaningful for temporary status"`
	Files       []string `json:"files,omitempty" jsonschema:"repo-relative file paths the entry refers to"`
	Body        string   `json:"body" jsonschema:"Markdown body. Include sub-headings like ### Decision/Reason/Alternatives or ### Problem/Solution/Example as fits the entry type."`
}

// SaveResult returns the persisted entry's id so the agent can echo
// it back to the user or reference it in a follow-up tool call.
type SaveResult struct {
	ID      string `json:"id"`
	Section string `json:"section"`
}

func handleSaveDecision(paths brain.BrainPaths) func(context.Context, *mcp.CallToolRequest, SaveArgs) (*mcp.CallToolResult, SaveResult, error) {
	return makeSaveHandler(paths, brain.SaveTypeDecision, "decisions")
}

func handleSaveBugfix(paths brain.BrainPaths) func(context.Context, *mcp.CallToolRequest, SaveArgs) (*mcp.CallToolResult, SaveResult, error) {
	return makeSaveHandler(paths, brain.SaveTypeBugfix, "bugfixes")
}

func handleSaveTesting(paths brain.BrainPaths) func(context.Context, *mcp.CallToolRequest, SaveArgs) (*mcp.CallToolResult, SaveResult, error) {
	return makeSaveHandler(paths, brain.SaveTypeTesting, "testing")
}

// makeSaveHandler is shared across the three save_* tools — every save
// goes through brain.AppendEntry, which enforces the brain-initialized
// guard, validates the entry, and writes atomically.
func makeSaveHandler(paths brain.BrainPaths, saveType brain.SaveType, section string) func(context.Context, *mcp.CallToolRequest, SaveArgs) (*mcp.CallToolResult, SaveResult, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, args SaveArgs) (*mcp.CallToolResult, SaveResult, error) {
		entry := &brain.SaveEntry{
			Type:        saveType,
			Title:       args.Title,
			Platform:    args.Platform,
			Area:        args.Area,
			Status:      args.Status,
			ReviewAfter: args.ReviewAfter,
			Files:       args.Files,
			Body:        args.Body,
		}
		id, err := brain.AppendEntry(paths, entry)
		if err != nil {
			return nil, SaveResult{}, err
		}
		text := fmt.Sprintf("✓ saved to %s.md (id: %s)", section, id)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, SaveResult{ID: id, Section: section}, nil
	}
}

// brainNotInitialized returns the standard "init me first" error in
// the same Spanish phrasing the CLI uses. Keeping a single source of
// truth here means MCP clients and CLI users see the same message.
func brainNotInitialized(paths brain.BrainPaths) error {
	return fmt.Errorf("brain not initialized at %s — run `mobiai brain init` first", paths.Root)
}
