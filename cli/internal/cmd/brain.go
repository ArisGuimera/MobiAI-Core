package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/brain"
	brainmcp "github.com/ArisGuimera/MobiAI-Core/cli/internal/brain/mcp"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/branding"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/i18n"
)

// NewBrainCmd builds `mobiai brain <init|scan|context>`.
//
// Phase 1 of MobiAI Brain — a per-project memory layer for mobile agents.
// All state lives inside the project at <root>/.mobiai/brain/, never in
// the global ~/.mobiai/ directory (that's the CLI's own state).
func NewBrainCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "brain",
		Short: i18n.T("Per-project memory for mobile agents"),
		Long: i18n.T("MobiAI Brain stores living per-project context: detected stack, decisions, bugfixes, testing patterns and integrations. Lives at <repo>/.mobiai/brain/ — separate from the CLI's global state."),
	}
	root.AddCommand(
		newBrainInitCmd(),
		newBrainScanCmd(),
		newBrainContextCmd(),
		newBrainSaveCmd(),
		newBrainSearchCmd(),
		newBrainReviewCmd(),
		newBrainPromoteCmd(),
		newBrainBumpCmd(),
		newBrainMCPCmd(),
		newBrainInstallMCPCmd(),
	)
	return root
}

// brainCommonFlags wires up the shared --root flag on a leaf command.
// When omitted, the command resolves the project root by walking up
// from the current working directory (see brain.FindProjectRoot).
func brainCommonFlags(c *cobra.Command) {
	c.Flags().String("root", "", i18n.T("project path (default: detected from cwd)"))
}

// resolveBrainRoot returns the project root for a brain command. It
// honors --root if given; otherwise it walks up from cwd looking for
// project markers (see brain.FindProjectRoot).
func resolveBrainRoot(c *cobra.Command) (brain.RootInfo, error) {
	if explicit, _ := c.Flags().GetString("root"); explicit != "" {
		abs, err := filepath.Abs(explicit)
		if err != nil {
			return brain.RootInfo{}, fmt.Errorf("resolve --root: %w", err)
		}
		if info, err := os.Stat(abs); err != nil || !info.IsDir() {
			return brain.RootInfo{}, fmt.Errorf("--root is not a directory: %s", abs)
		}
		return brain.RootInfo{Path: abs, Source: brain.RootSourceCwd}, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return brain.RootInfo{}, fmt.Errorf("getwd: %w", err)
	}
	return brain.FindProjectRoot(cwd)
}

func newBrainInitCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "init",
		Short: i18n.T("Initialize .mobiai/brain in the current project"),
		RunE:  runBrainInit,
	}
	brainCommonFlags(c)
	return c
}

func runBrainInit(c *cobra.Command, _ []string) error {
	out := c.OutOrStdout()
	// Banner first — `brain init` is the user's entry point into the
	// Brain ecosystem, so it's the right moment to show the wordmark.
	// Print() handles --no-color, NO_COLOR env, and TTY detection
	// internally; we just pass the flag value.
	flags := FlagsFromCmd(c)
	branding.Print(out, flags.NoColor)
	fmt.Fprintln(out, "")

	info, err := resolveBrainRoot(c)
	if err != nil {
		return err
	}
	if info.Warning != "" {
		fmt.Fprintf(out, i18n.T("⚠ %s (using %s)\n"), info.Warning, info.Path)
	} else {
		fmt.Fprintf(out, i18n.T("Detected project: %s\n"), info.Path)
	}

	paths := brain.NewBrainPaths(info.Path)
	if err := paths.EnsureDirs(); err != nil {
		return err
	}

	createdConfig := false
	if _, err := os.Stat(paths.ConfigFile); os.IsNotExist(err) {
		cfg := brain.DefaultConfig(info.Path)
		if err := cfg.Save(paths); err != nil {
			return err
		}
		createdConfig = true
	} else if err != nil {
		return fmt.Errorf("stat config.json: %w", err)
	}

	createdMemories, err := brain.EnsureMemoryFiles(paths)
	if err != nil {
		return err
	}

	if createdConfig {
		fmt.Fprintf(out, i18n.T("✓ created %s\n"), relTo(info.Path, paths.ConfigFile))
	} else {
		fmt.Fprintf(out, i18n.T("= %s already exists (not overwritten)\n"), relTo(info.Path, paths.ConfigFile))
	}
	if len(createdMemories) > 0 {
		fmt.Fprintf(out, i18n.T("✓ memories initialized: %s\n"), strings.Join(createdMemories, ", "))
	} else {
		fmt.Fprintln(out, i18n.T("= memories already exist (not overwritten)"))
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, i18n.T("Next step: `mobiai brain scan` to detect the stack."))
	return nil
}

func newBrainScanCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "scan",
		Short: i18n.T("Scan the project and detect the mobile stack"),
		RunE:  runBrainScan,
	}
	brainCommonFlags(c)
	return c
}

func runBrainScan(c *cobra.Command, _ []string) error {
	out := c.OutOrStdout()
	info, err := resolveBrainRoot(c)
	if err != nil {
		return err
	}
	paths := brain.NewBrainPaths(info.Path)
	if !paths.Exists() {
		return fmt.Errorf("could not find .mobiai/brain/config.json in %s. Run `mobiai brain init` first", info.Path)
	}

	fmt.Fprintf(out, i18n.T("Scanning %s ...\n"), info.Path)
	scan, err := brain.ScanProject(info.Path)
	if err != nil {
		return err
	}

	// Reflect the scan result back into config.json so subsequent
	// `brain context` runs see the freshest project_type and platforms
	// without forcing the user to edit the file by hand.
	cfg, err := brain.LoadConfig(paths)
	if err != nil {
		return fmt.Errorf("read config.json: %w", err)
	}
	if cfg.ProjectType == brain.ProjectTypeUnknown {
		cfg.ProjectType = scan.ProjectType
	}
	if len(cfg.Platforms) == 0 {
		cfg.Platforms = append([]string(nil), scan.Platforms...)
	}
	if err := cfg.Save(paths); err != nil {
		return err
	}
	if err := scan.Save(paths); err != nil {
		return err
	}

	printScanSummary(out, scan)
	fmt.Fprintf(out, i18n.T("✓ scan saved to %s\n"), relTo(info.Path, paths.ScanFile))
	return nil
}

func printScanSummary(out io.Writer, s *brain.Scan) {
	w := func(format string, args ...interface{}) {
		fmt.Fprintf(out, format, args...)
	}
	fmt.Fprint(out, i18n.T("\nSummary:\n"))
	w(i18n.T("  Type:        %s\n"), s.ProjectType)
	if len(s.Platforms) > 0 {
		w(i18n.T("  Platforms:   %s\n"), strings.Join(s.Platforms, ", "))
	}
	for _, row := range []struct {
		label string
		items []string
	}{
		{"UI", s.UI},
		{"DI", s.DI},
		{"Network", s.Network},
		{"Persistence", s.Persistence},
		{"Testing", s.Testing},
		{"Integrations", s.Integrations},
		{"CI/CD", s.CICD},
	} {
		if len(row.items) == 0 {
			continue
		}
		w("  %-12s %s\n", row.label+":", strings.Join(row.items, ", "))
	}
	if len(s.Warnings) > 0 {
		w(i18n.T("  Warnings:    %d\n"), len(s.Warnings))
	}
	w("\n")
}

func newBrainContextCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "context",
		Short: i18n.T("Print the Brain's Markdown context (config + scan + memories)"),
		Long: i18n.T("Prints the Brain context as Markdown. Without flags, dumps everything. With --section, --platform, --status or --area, filters memory entries so only what's relevant appears (useful as the brain grows)."),
		RunE: runBrainContext,
	}
	brainCommonFlags(c)
	c.Flags().StringSlice("section", nil,
		i18n.T("limit sections to render (stack,rules,decisions,bugfixes,testing,integrations,releases,warnings). Repeatable or comma-separated."))
	c.Flags().String("platform", "",
		i18n.T("filter entries by platform (android|ios|shared|kmp|flutter|react-native)"))
	c.Flags().String("status", "",
		i18n.T("filter entries by status (active|temporary|deprecated)"))
	c.Flags().String("area", "",
		i18n.T("filter entries whose area contains this string (substring)"))
	return c
}

func runBrainContext(c *cobra.Command, _ []string) error {
	info, err := resolveBrainRoot(c)
	if err != nil {
		return err
	}
	paths := brain.NewBrainPaths(info.Path)
	if !paths.Exists() {
		return fmt.Errorf("could not find .mobiai/brain/config.json in %s. Run `mobiai brain init` first", info.Path)
	}
	cfg, err := brain.LoadConfig(paths)
	if err != nil {
		return err
	}
	scan, err := brain.LoadScan(paths)
	if err != nil {
		return err
	}
	sections, _ := c.Flags().GetStringSlice("section")
	platform, _ := c.Flags().GetString("platform")
	status, _ := c.Flags().GetString("status")
	area, _ := c.Flags().GetString("area")
	md := brain.BuildContextWith(cfg, scan, paths, brain.ContextOptions{
		Sections: sections,
		Filter: brain.EntryFilter{
			Platform: platform,
			Status:   status,
			Area:     area,
		},
	})
	fmt.Fprint(c.OutOrStdout(), md)
	return nil
}

func newBrainSearchCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "search <query>",
		Short: i18n.T("Free-text search across Brain memories"),
		Long: i18n.T("Performs a case-insensitive match against the title and body of each memory entry. The --platform/--status/--area flags filter results with AND semantics."),
		Args: cobra.MinimumNArgs(1),
		RunE: runBrainSearch,
	}
	brainCommonFlags(c)
	c.Flags().String("platform", "", i18n.T("limit to entries with this platform"))
	c.Flags().String("status", "", i18n.T("limit to entries with this status"))
	c.Flags().String("area", "", i18n.T("limit to entries whose area contains this string"))
	return c
}

func runBrainSearch(c *cobra.Command, args []string) error {
	out := c.OutOrStdout()
	info, err := resolveBrainRoot(c)
	if err != nil {
		return err
	}
	paths := brain.NewBrainPaths(info.Path)
	if !paths.Exists() {
		return fmt.Errorf("brain not initialized in %s — run `mobiai brain init` first", info.Path)
	}
	query := strings.Join(args, " ")
	platform, _ := c.Flags().GetString("platform")
	status, _ := c.Flags().GetString("status")
	area, _ := c.Flags().GetString("area")
	hits, err := brain.Search(paths, query, brain.EntryFilter{
		Platform: platform,
		Status:   status,
		Area:     area,
	})
	if err != nil {
		return err
	}
	if len(hits) == 0 {
		fmt.Fprintf(out, i18n.T("No results for %q.\n"), query)
		return nil
	}
	fmt.Fprintf(out, i18n.T("%d result(s) for %q:\n\n"), len(hits), query)
	for _, h := range hits {
		// Format: `[section] title (status, platform) — snippet`
		var meta []string
		if s := h.Entry.Get("status"); s != "" {
			meta = append(meta, s)
		}
		if p := h.Entry.Get("platform"); p != "" {
			meta = append(meta, p)
		}
		suffix := ""
		if len(meta) > 0 {
			suffix = " (" + strings.Join(meta, ", ") + ")"
		}
		fmt.Fprintf(out, "[%s] %s%s\n", h.Section, h.Entry.Title, suffix)
		if h.Snippet != "" {
			fmt.Fprintf(out, "    %s\n", truncate(h.Snippet, 120))
		}
		if id := h.Entry.Get("id"); id != "" {
			fmt.Fprintf(out, "    id: %s\n", id)
		}
		fmt.Fprintln(out)
	}
	return nil
}

// truncate cuts s at maxLen with an ellipsis, avoiding a slice on a
// multi-byte rune boundary.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}

func newBrainReviewCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "review",
		Short: i18n.T("List temporary entries whose review_after has passed"),
		Long: i18n.T("Walks the memories and shows entries with status: temporary whose review_after <= today. Designed to keep temporary workarounds from drifting into permanence by inertia. By default exits with status 1 if there are overdue entries (useful as a CI / pre-commit gate); pass --no-fail to only report. With --include-no-date it also lists temporary entries without a review_after assigned."),
		RunE: runBrainReview,
	}
	brainCommonFlags(c)
	c.Flags().Bool("include-no-date", false,
		i18n.T("also list temporary entries without review_after (printed under a separate heading)"))
	c.Flags().Bool("no-fail", false,
		i18n.T("always exit 0, even if there are overdue entries (informational mode)"))
	return c
}

func runBrainReview(c *cobra.Command, _ []string) error {
	out := c.OutOrStdout()
	info, err := resolveBrainRoot(c)
	if err != nil {
		return err
	}
	paths := brain.NewBrainPaths(info.Path)
	if !paths.Exists() {
		return fmt.Errorf("brain not initialized in %s — run `mobiai brain init` first", info.Path)
	}

	includeNoDate, _ := c.Flags().GetBool("include-no-date")
	noFail, _ := c.Flags().GetBool("no-fail")

	items, err := brain.Review(paths, brain.ReviewOptions{
		IncludeNoDate: includeNoDate,
	})
	if err != nil {
		return err
	}

	overdue, noDate := splitReviewItems(items)

	if len(overdue) == 0 && len(noDate) == 0 {
		fmt.Fprintln(out, i18n.T("✓ No overdue temporary entries."))
		return nil
	}

	if len(overdue) > 0 {
		fmt.Fprintf(out, i18n.T("⚠ %d overdue temporary entry(ies):\n\n"), len(overdue))
		printOverdueItems(out, overdue)
	} else {
		fmt.Fprintln(out, i18n.T("✓ No overdue entries."))
	}

	if len(noDate) > 0 {
		fmt.Fprintf(out, i18n.T("\n%d temporary entry(ies) without review_after:\n\n"), len(noDate))
		printNoDateItems(out, noDate)
	}

	// Non-zero exit when there's actually overdue debt and the caller
	// didn't opt out. Useful as a CI gate. os.Exit is fine here — no
	// deferred cleanup in this command. Tests should call brain.Review
	// directly to avoid bypassing the test runner.
	if len(overdue) > 0 && !noFail {
		os.Exit(1)
	}
	return nil
}

// splitReviewItems separates dated/overdue entries from no-date ones so
// the CLI can render them under different headings.
func splitReviewItems(items []brain.ReviewItem) (overdue, noDate []brain.ReviewItem) {
	for _, it := range items {
		if it.HasDate {
			overdue = append(overdue, it)
		} else {
			noDate = append(noDate, it)
		}
	}
	return overdue, noDate
}

func printOverdueItems(out io.Writer, items []brain.ReviewItem) {
	currentSection := ""
	for _, it := range items {
		if it.Section != currentSection {
			fmt.Fprintf(out, "%s.md\n", it.Section)
			currentSection = it.Section
		}
		fmt.Fprintf(out, "  ⚠ %s\n", it.Entry.Title)
		fmt.Fprintf(out, i18n.T("    review_after: %s (overdue by %s)\n"),
			it.ReviewAfter, daysOverdueLabel(it.DaysOverdue))
		if p := it.Entry.Get("platform"); p != "" {
			fmt.Fprintf(out, "    platform: %s\n", p)
		}
		if a := it.Entry.Get("area"); a != "" {
			fmt.Fprintf(out, "    area: %s\n", a)
		}
		if id := it.Entry.Get("id"); id != "" {
			fmt.Fprintf(out, "    id: %s\n", id)
		}
		fmt.Fprintln(out)
	}
}

func printNoDateItems(out io.Writer, items []brain.ReviewItem) {
	currentSection := ""
	for _, it := range items {
		if it.Section != currentSection {
			fmt.Fprintf(out, "%s.md\n", it.Section)
			currentSection = it.Section
		}
		fmt.Fprintf(out, "  • %s\n", it.Entry.Title)
		if p := it.Entry.Get("platform"); p != "" {
			fmt.Fprintf(out, "    platform: %s\n", p)
		}
		if id := it.Entry.Get("id"); id != "" {
			fmt.Fprintf(out, "    id: %s\n", id)
		}
		fmt.Fprintln(out)
	}
}

func newBrainPromoteCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "promote <id>",
		Short: i18n.T("Change the status of an existing entry"),
		Long: i18n.T("Updates the status: of the entry with the given id (active|temporary|deprecated). Designed as the wrap-up flow after `brain review`: once you no longer need a temporary entry, promote it to active (it became permanent) or deprecated (no longer applies). Optionally also updates review_after in the same call, or removes it with --clear-review-after."),
		Args: cobra.ExactArgs(1),
		RunE: runBrainPromote,
	}
	brainCommonFlags(c)
	c.Flags().String("status", "", i18n.T("new status: active|temporary|deprecated (required)"))
	c.Flags().String("review-after", "", i18n.T("also update review_after to this YYYY-MM-DD date (optional)"))
	c.Flags().Bool("clear-review-after", false, i18n.T("remove review_after (incompatible with --review-after)"))
	_ = c.MarkFlagRequired("status")
	return c
}

func runBrainPromote(c *cobra.Command, args []string) error {
	id := args[0]
	status, _ := c.Flags().GetString("status")
	reviewAfter, _ := c.Flags().GetString("review-after")
	clearReview, _ := c.Flags().GetBool("clear-review-after")
	return doBrainUpdate(c, id, brain.UpdateOptions{
		Status:           status,
		ReviewAfter:      reviewAfter,
		ClearReviewAfter: clearReview,
	})
}

func newBrainBumpCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "bump <id>",
		Short: i18n.T("Extend the review_after of an existing entry"),
		Long: i18n.T("Updates review_after of the entry with the given id, without touching status. Useful after `brain review` when a temporary entry is still valid and you want to extend its review deadline. For status changes use `promote`."),
		Args: cobra.ExactArgs(1),
		RunE: runBrainBump,
	}
	brainCommonFlags(c)
	c.Flags().String("review-after", "", i18n.T("new YYYY-MM-DD date for review_after (required)"))
	_ = c.MarkFlagRequired("review-after")
	return c
}

func runBrainBump(c *cobra.Command, args []string) error {
	id := args[0]
	reviewAfter, _ := c.Flags().GetString("review-after")
	return doBrainUpdate(c, id, brain.UpdateOptions{
		ReviewAfter: reviewAfter,
	})
}

// doBrainUpdate is the shared backend for promote/bump. Both commands
// are thin wrappers over UpdateEntry — the difference is which flags
// they accept, not what happens on disk.
func doBrainUpdate(c *cobra.Command, id string, opts brain.UpdateOptions) error {
	out := c.OutOrStdout()
	info, err := resolveBrainRoot(c)
	if err != nil {
		return err
	}
	paths := brain.NewBrainPaths(info.Path)
	if !paths.Exists() {
		return fmt.Errorf("brain not initialized in %s — run `mobiai brain init` first", info.Path)
	}
	res, err := brain.UpdateEntry(paths, id, opts)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, i18n.T("✓ %s updated (%s)\n"), res.Title, res.File)
	if res.PrevStatus != res.NewStatus {
		fmt.Fprintf(out, "  status: %s → %s\n", res.PrevStatus, res.NewStatus)
	}
	if res.PrevReviewAfter != res.NewReviewAfter {
		from := res.PrevReviewAfter
		if from == "" {
			from = "(none)"
		}
		to := res.NewReviewAfter
		if to == "" {
			to = "(none)"
		}
		fmt.Fprintf(out, "  review_after: %s → %s\n", from, to)
	}
	return nil
}

// daysOverdueLabel renders a human-readable "N days" / "1 day" / "today"
// label for a non-negative number of days overdue.
func daysOverdueLabel(days int) string {
	switch {
	case days == 0:
		return i18n.T("today")
	case days == 1:
		return i18n.T("1 day")
	default:
		return fmt.Sprintf(i18n.T("%d days"), days)
	}
}

func newBrainSaveCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "save",
		Short: i18n.T("Save an entry into the Brain's memories"),
		Long: i18n.T("Appends a structured entry to the matching memory file. Requires the Brain to be initialized in the project (run `mobiai brain init` first)."),
	}
	root.AddCommand(
		newBrainSaveSubCmd(brain.SaveTypeDecision, "decision",
			i18n.T("Save an architecture decision")),
		newBrainSaveSubCmd(brain.SaveTypeBugfix, "bugfix",
			i18n.T("Save a bugfix or workaround")),
		newBrainSaveSubCmd(brain.SaveTypeTesting, "testing",
			i18n.T("Save a reusable testing pattern")),
	)
	return root
}

func newBrainSaveSubCmd(saveType brain.SaveType, use, short string) *cobra.Command {
	c := &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runBrainSave(cmd, saveType)
		},
	}
	brainCommonFlags(c)
	c.Flags().String("title", "", i18n.T("short entry title (required)"))
	c.Flags().String("platform", "", i18n.T("android|ios|shared|kmp|flutter|react-native (optional)"))
	c.Flags().String("area", "", i18n.T("project area (free-form, optional)"))
	c.Flags().String("status", "active", i18n.T("active|temporary|deprecated"))
	c.Flags().String("review-after", "", i18n.T("YYYY-MM-DD date to review (optional)"))
	c.Flags().String("body", "", i18n.T("Markdown body (if omitted, read from stdin)"))
	c.Flags().StringSlice("files", nil, i18n.T("related files, comma-separated (optional)"))
	_ = c.MarkFlagRequired("title")
	return c
}

func runBrainSave(c *cobra.Command, saveType brain.SaveType) error {
	out := c.OutOrStdout()
	info, err := resolveBrainRoot(c)
	if err != nil {
		return err
	}
	paths := brain.NewBrainPaths(info.Path)
	if !paths.Exists() {
		return fmt.Errorf("brain not initialized in %s — run `mobiai brain init` first", info.Path)
	}

	title, _ := c.Flags().GetString("title")
	platform, _ := c.Flags().GetString("platform")
	area, _ := c.Flags().GetString("area")
	status, _ := c.Flags().GetString("status")
	reviewAfter, _ := c.Flags().GetString("review-after")
	body, _ := c.Flags().GetString("body")
	files, _ := c.Flags().GetStringSlice("files")

	if body == "" {
		// stdin fallback: lets the agent pipe a multi-line markdown body
		// without struggling with shell-escaping a flag value.
		piped, err := readPipedStdin(c.InOrStdin())
		if err != nil {
			return fmt.Errorf("read body from stdin: %w", err)
		}
		body = piped
	}

	entry := &brain.SaveEntry{
		Type:        saveType,
		Title:       title,
		Platform:    platform,
		Area:        area,
		Status:      status,
		ReviewAfter: reviewAfter,
		Body:        body,
		Files:       files,
	}
	id, err := brain.AppendEntry(paths, entry)
	if err != nil {
		return err
	}

	fileBase := map[brain.SaveType]string{
		brain.SaveTypeDecision: "decisions.md",
		brain.SaveTypeBugfix:   "bugfixes.md",
		brain.SaveTypeTesting:  "testing.md",
	}[saveType]
	rel := filepath.Join(".mobiai", "brain", "memories", fileBase)
	fmt.Fprintf(out, i18n.T("✓ saved to %s\n  id: %s\n"), rel, id)
	return nil
}

// readPipedStdin returns stdin content only when stdin is a pipe (not a
// tty). Returns "" when running interactively so the command doesn't
// hang waiting for input the user didn't intend to provide.
func readPipedStdin(in io.Reader) (string, error) {
	f, ok := in.(*os.File)
	if ok {
		stat, err := f.Stat()
		if err != nil {
			return "", err
		}
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			// stdin is a terminal — nothing being piped.
			return "", nil
		}
	}
	var b strings.Builder
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		b.WriteString(scanner.Text())
		b.WriteString("\n")
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func newBrainMCPCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "mcp",
		Short: i18n.T("Start an MCP server that exposes the Brain as tools"),
		Long: i18n.T("Starts an MCP (Model Context Protocol) server that exposes the Brain's operations (context, search, scan, save) as tools the agent can invoke directly. Communicates over stdio — designed for an MCP client (Claude Code, Cursor, Copilot CLI, etc.) to launch it as a subprocess. See brain/MCP-SETUP.md."),
		RunE: runBrainMCP,
	}
	brainCommonFlags(c)
	return c
}

func runBrainMCP(c *cobra.Command, _ []string) error {
	info, err := resolveBrainRoot(c)
	if err != nil {
		return err
	}
	server, err := brainmcp.NewServer(info.Path, brainCLIVersion())
	if err != nil {
		return err
	}
	// Run blocks until the MCP client disconnects. Use cmd's context so
	// shell signals (SIGINT, etc.) terminate the server cleanly.
	return server.Run(c.Context(), &mcp.StdioTransport{})
}

func newBrainInstallMCPCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "install-mcp",
		Short: i18n.T("Register the Brain MCP server in AI clients (Claude Code, Cursor)"),
		Long: i18n.T("Adds the `mobiai-brain` server to each supported AI client's config, preserving the rest of the file. By default detects which clients are present (presence of ~/.claude or ~/.cursor); use --client to force a single one. Idempotent: re-running it with the same config is a no-op. Use --dry-run to preview without touching files, or --uninstall to remove the registration."),
		RunE: runBrainInstallMCP,
	}
	c.Flags().StringSlice("client", nil,
		i18n.T("clients to register: claude|cursor (repeatable or comma-separated; default: all detected)"))
	c.Flags().Bool("dry-run", false, i18n.T("show which files would be touched without writing anything"))
	c.Flags().Bool("uninstall", false, i18n.T("remove the registration instead of creating it"))
	c.Flags().String("binary", "", i18n.T("path to the mobiai binary to register (default: the binary in use)"))
	return c
}

func runBrainInstallMCP(c *cobra.Command, _ []string) error {
	out := c.OutOrStdout()
	rawClients, _ := c.Flags().GetStringSlice("client")
	dryRun, _ := c.Flags().GetBool("dry-run")
	uninstall, _ := c.Flags().GetBool("uninstall")
	binary, _ := c.Flags().GetString("binary")

	clients, err := parseMCPClients(rawClients)
	if err != nil {
		return err
	}

	opts := brain.InstallOptions{
		Clients:    clients,
		BinaryPath: binary,
		DryRun:     dryRun,
	}

	var results []brain.InstallResult
	if uninstall {
		results, err = brain.UninstallMCP(opts)
	} else {
		results, err = brain.InstallMCP(opts)
	}
	if err != nil {
		return err
	}

	printInstallResults(out, results, uninstall)
	return nil
}

// parseMCPClients turns the raw --client flag values into typed clients.
// Empty input means "let InstallMCP pick defaults" (all supported).
func parseMCPClients(raw []string) ([]brain.MCPClient, error) {
	out := make([]brain.MCPClient, 0, len(raw))
	for _, r := range raw {
		switch strings.ToLower(strings.TrimSpace(r)) {
		case "claude", "claude-code", "claudecode":
			out = append(out, brain.MCPClientClaudeCode)
		case "cursor":
			out = append(out, brain.MCPClientCursor)
		case "":
			// Skip empties from accidental ",,".
			continue
		default:
			return nil, fmt.Errorf("--client %q unknown (supported: claude, cursor)", r)
		}
	}
	return out, nil
}

// printInstallResults renders one line per client with the action
// classification and config path. Suffixes "(dry-run)" so users
// don't think actions were applied when they weren't.
func printInstallResults(out io.Writer, results []brain.InstallResult, uninstall bool) {
	anyTouched := false
	for _, r := range results {
		icon, label := installIconAndLabel(r.Action)
		suffix := ""
		if r.DryRun {
			suffix = " (dry-run)"
		}
		fmt.Fprintf(out, "%s [%s] %s — %s%s\n", icon, r.Client, label, r.ConfigPath, suffix)
		if r.Action == brain.ActionInstalled || r.Action == brain.ActionUpdated || r.Action == brain.ActionUninstalled {
			anyTouched = true
		}
	}
	if anyTouched && !results[0].DryRun {
		fmt.Fprintln(out, "")
		if uninstall {
			fmt.Fprintln(out, i18n.T("Done. Restart the client for changes to take effect."))
		} else {
			fmt.Fprintln(out, i18n.T("Done. Restart the client so it loads the MCP server."))
		}
	}
}

func installIconAndLabel(a brain.InstallAction) (string, string) {
	switch a {
	case brain.ActionInstalled:
		return "✓", i18n.T("registered")
	case brain.ActionUpdated:
		return "✓", i18n.T("updated")
	case brain.ActionUnchanged:
		return "=", i18n.T("unchanged (already registered)")
	case brain.ActionUninstalled:
		return "✓", i18n.T("removed")
	case brain.ActionNotPresent:
		return "=", i18n.T("was not registered")
	case brain.ActionSkipped:
		return "·", i18n.T("skipped (client not installed)")
	}
	return "?", string(a)
}

// brainCLIVersion returns the version the CLI was built with, falling
// back to "dev" when unset. Used to advertise the server version to
// MCP clients. Reuses the package-level `version` populated by
// SetVersion (see status.go).
func brainCLIVersion() string {
	if version == "" {
		return "dev"
	}
	return version
}

// relTo returns target relative to base, falling back to the absolute
// path when filepath.Rel fails (e.g. on different volumes).
func relTo(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}
