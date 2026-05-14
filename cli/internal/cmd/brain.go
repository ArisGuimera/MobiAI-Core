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
)

// NewBrainCmd builds `mobiai brain <init|scan|context>`.
//
// Phase 1 of MobiAI Brain — a per-project memory layer for mobile agents.
// All state lives inside the project at <root>/.mobiai/brain/, never in
// the global ~/.mobiai/ directory (that's the CLI's own state).
func NewBrainCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "brain",
		Short: "Memoria por proyecto para agentes mobile",
		Long: "MobiAI Brain guarda contexto vivo por proyecto: stack detectado, " +
			"decisiones, bugfixes, patrones de testing e integraciones. " +
			"Vive en <repo>/.mobiai/brain/ — separado del estado global de la CLI.",
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
	)
	return root
}

// brainCommonFlags wires up the shared --root flag on a leaf command.
// When omitted, the command resolves the project root by walking up
// from the current working directory (see brain.FindProjectRoot).
func brainCommonFlags(c *cobra.Command) {
	c.Flags().String("root", "", "ruta del proyecto (default: detectado desde el cwd)")
}

// resolveBrainRoot returns the project root for a brain command. It
// honors --root if given; otherwise it walks up from cwd looking for
// project markers (see brain.FindProjectRoot).
func resolveBrainRoot(c *cobra.Command) (brain.RootInfo, error) {
	if explicit, _ := c.Flags().GetString("root"); explicit != "" {
		abs, err := filepath.Abs(explicit)
		if err != nil {
			return brain.RootInfo{}, fmt.Errorf("absolutizar --root: %w", err)
		}
		if info, err := os.Stat(abs); err != nil || !info.IsDir() {
			return brain.RootInfo{}, fmt.Errorf("--root no es un directorio: %s", abs)
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
		Short: "Inicializa .mobiai/brain en el proyecto actual",
		RunE:  runBrainInit,
	}
	brainCommonFlags(c)
	return c
}

func runBrainInit(c *cobra.Command, _ []string) error {
	out := c.OutOrStdout()
	info, err := resolveBrainRoot(c)
	if err != nil {
		return err
	}
	if info.Warning != "" {
		fmt.Fprintf(out, "⚠ %s (usando %s)\n", info.Warning, info.Path)
	} else {
		fmt.Fprintf(out, "Proyecto detectado: %s\n", info.Path)
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
		fmt.Fprintf(out, "✓ creado %s\n", relTo(info.Path, paths.ConfigFile))
	} else {
		fmt.Fprintf(out, "= %s ya existe (no se sobrescribe)\n", relTo(info.Path, paths.ConfigFile))
	}
	if len(createdMemories) > 0 {
		fmt.Fprintf(out, "✓ memorias inicializadas: %s\n", strings.Join(createdMemories, ", "))
	} else {
		fmt.Fprintln(out, "= memorias ya existentes (no se sobrescriben)")
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Siguiente paso: `mobiai brain scan` para detectar el stack.")
	return nil
}

func newBrainScanCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "scan",
		Short: "Escanea el proyecto y detecta el stack mobile",
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
		return fmt.Errorf("no encontré .mobiai/brain/config.json en %s. Corré `mobiai brain init` primero", info.Path)
	}

	fmt.Fprintf(out, "Escaneando %s ...\n", info.Path)
	scan, err := brain.ScanProject(info.Path)
	if err != nil {
		return err
	}

	// Reflect the scan result back into config.json so subsequent
	// `brain context` runs see the freshest project_type and platforms
	// without forcing the user to edit the file by hand.
	cfg, err := brain.LoadConfig(paths)
	if err != nil {
		return fmt.Errorf("leer config.json: %w", err)
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
	fmt.Fprintf(out, "✓ scan guardado en %s\n", relTo(info.Path, paths.ScanFile))
	return nil
}

func printScanSummary(out io.Writer, s *brain.Scan) {
	w := func(format string, args ...interface{}) {
		fmt.Fprintf(out, format, args...)
	}
	w("\nResumen:\n")
	w("  Tipo:        %s\n", s.ProjectType)
	if len(s.Platforms) > 0 {
		w("  Plataformas: %s\n", strings.Join(s.Platforms, ", "))
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
		w("  Warnings:    %d\n", len(s.Warnings))
	}
	w("\n")
}

func newBrainContextCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "context",
		Short: "Imprime el contexto Markdown del Brain (config + scan + memorias)",
		Long: "Imprime el contexto del Brain como Markdown. Sin flags, vuelca todo. " +
			"Con --section, --platform, --status o --area filtra entradas de las " +
			"memorias para que solo aparezca lo relevante (útil cuando el brain crece).",
		RunE: runBrainContext,
	}
	brainCommonFlags(c)
	c.Flags().StringSlice("section", nil,
		"limita las secciones a renderizar (stack,rules,decisions,bugfixes,testing,integrations,releases,warnings). Repetible o coma-separado.")
	c.Flags().String("platform", "",
		"filtra entradas por platform (android|ios|shared|kmp|flutter|react-native)")
	c.Flags().String("status", "",
		"filtra entradas por status (active|temporary|deprecated)")
	c.Flags().String("area", "",
		"filtra entradas cuyo area contenga este string (substring)")
	return c
}

func runBrainContext(c *cobra.Command, _ []string) error {
	info, err := resolveBrainRoot(c)
	if err != nil {
		return err
	}
	paths := brain.NewBrainPaths(info.Path)
	if !paths.Exists() {
		return fmt.Errorf("no encontré .mobiai/brain/config.json en %s. Corré `mobiai brain init` primero", info.Path)
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
		Short: "Busca texto libre en las memorias del Brain",
		Long: "Hace match case-insensitive sobre el título y el cuerpo de cada " +
			"entrada de las memorias. Los flags --platform/--status/--area filtran " +
			"los resultados con semántica AND.",
		Args: cobra.MinimumNArgs(1),
		RunE: runBrainSearch,
	}
	brainCommonFlags(c)
	c.Flags().String("platform", "", "limita a entradas con este platform")
	c.Flags().String("status", "", "limita a entradas con este status")
	c.Flags().String("area", "", "limita a entradas cuya area contenga este string")
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
		return fmt.Errorf("brain no inicializado en %s — corré `mobiai brain init` primero", info.Path)
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
		fmt.Fprintf(out, "Sin resultados para %q.\n", query)
		return nil
	}
	fmt.Fprintf(out, "%d resultado(s) para %q:\n\n", len(hits), query)
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
		Short: "Lista entradas temporales cuyo review_after ya pasó",
		Long: "Recorre las memorias y muestra las entradas con status: temporary " +
			"cuyo review_after <= hoy. Pensado para evitar que workarounds temporales " +
			"se vuelvan permanentes por inercia. Por defecto sale con exit 1 si hay " +
			"vencidas (útil como gate en CI / pre-commit); usá --no-fail para que " +
			"solo informe. Con --include-no-date también lista las temporary sin " +
			"review_after asignado.",
		RunE: runBrainReview,
	}
	brainCommonFlags(c)
	c.Flags().Bool("include-no-date", false,
		"también lista entradas temporary sin review_after (sección aparte en la salida)")
	c.Flags().Bool("no-fail", false,
		"siempre exit 0, incluso si hay entradas vencidas (modo solo-informativo)")
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
		return fmt.Errorf("brain no inicializado en %s — corré `mobiai brain init` primero", info.Path)
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
		fmt.Fprintln(out, "✓ No hay entradas temporales vencidas.")
		return nil
	}

	if len(overdue) > 0 {
		fmt.Fprintf(out, "⚠ %d entrada(s) temporal(es) vencida(s):\n\n", len(overdue))
		printOverdueItems(out, overdue)
	} else {
		fmt.Fprintln(out, "✓ No hay entradas vencidas.")
	}

	if len(noDate) > 0 {
		fmt.Fprintf(out, "\n%d entrada(s) temporal(es) sin review_after:\n\n", len(noDate))
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
		fmt.Fprintf(out, "    review_after: %s (vencido hace %s)\n",
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
		Short: "Cambia el status de una entrada existente",
		Long: "Actualiza el status: de la entrada con el id dado (active|temporary|" +
			"deprecated). Pensado como flujo de cierre tras `brain review`: cuando ya no " +
			"necesitás una entrada temporary, la promovés a active (se volvió definitiva) " +
			"o deprecated (ya no aplica). Opcionalmente también actualiza review_after en " +
			"la misma llamada, o lo elimina con --clear-review-after.",
		Args: cobra.ExactArgs(1),
		RunE: runBrainPromote,
	}
	brainCommonFlags(c)
	c.Flags().String("status", "", "nuevo status: active|temporary|deprecated (requerido)")
	c.Flags().String("review-after", "", "actualizar también review_after a esta fecha YYYY-MM-DD (opcional)")
	c.Flags().Bool("clear-review-after", false, "eliminar review_after (incompatible con --review-after)")
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
		Short: "Extiende review_after de una entrada existente",
		Long: "Actualiza review_after de la entrada con el id dado, sin tocar status. " +
			"Útil tras `brain review` cuando una entrada temporary sigue siendo válida y " +
			"querés extender su plazo de revisión. Para cambios de status usá `promote`.",
		Args: cobra.ExactArgs(1),
		RunE: runBrainBump,
	}
	brainCommonFlags(c)
	c.Flags().String("review-after", "", "nueva fecha YYYY-MM-DD para review_after (requerido)")
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
		return fmt.Errorf("brain no inicializado en %s — corré `mobiai brain init` primero", info.Path)
	}
	res, err := brain.UpdateEntry(paths, id, opts)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "✓ %s actualizada (%s)\n", res.Title, res.File)
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

// daysOverdueLabel renders a human-readable "N días" / "N día" / "hoy"
// label for a non-negative number of days overdue.
func daysOverdueLabel(days int) string {
	switch {
	case days == 0:
		return "hoy"
	case days == 1:
		return "1 día"
	default:
		return fmt.Sprintf("%d días", days)
	}
}

func newBrainSaveCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "save",
		Short: "Guardar una entrada en las memorias del Brain",
		Long: "Añade una entrada estructurada al archivo de memoria correspondiente. " +
			"Requiere que el Brain esté inicializado en el proyecto " +
			"(corré `mobiai brain init` primero).",
	}
	root.AddCommand(
		newBrainSaveSubCmd(brain.SaveTypeDecision, "decision",
			"Guardar una decisión de arquitectura"),
		newBrainSaveSubCmd(brain.SaveTypeBugfix, "bugfix",
			"Guardar un bugfix o workaround"),
		newBrainSaveSubCmd(brain.SaveTypeTesting, "testing",
			"Guardar un patrón de testing reusable"),
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
	c.Flags().String("title", "", "título corto de la entrada (requerido)")
	c.Flags().String("platform", "", "android|ios|shared|kmp|flutter|react-native (opcional)")
	c.Flags().String("area", "", "área del proyecto (libre, opcional)")
	c.Flags().String("status", "active", "active|temporary|deprecated")
	c.Flags().String("review-after", "", "fecha YYYY-MM-DD para revisar (opcional)")
	c.Flags().String("body", "", "cuerpo Markdown (si se omite, se lee de stdin)")
	c.Flags().StringSlice("files", nil, "archivos relacionados, separados por coma (opcional)")
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
		return fmt.Errorf("brain no inicializado en %s — corré `mobiai brain init` primero", info.Path)
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
			return fmt.Errorf("leer body desde stdin: %w", err)
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
	fmt.Fprintf(out, "✓ guardado en %s\n  id: %s\n", rel, id)
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
		Short: "Arranca un servidor MCP que expone el Brain como tools",
		Long: "Arranca un servidor MCP (Model Context Protocol) que expone las operaciones " +
			"del Brain (context, search, scan, save) como tools que el agente puede invocar " +
			"directamente. Comunica por stdio — pensado para que un cliente MCP (Claude Code, " +
			"Cursor, Copilot CLI, etc.) lo lance como subproceso. Ver brain/MCP-SETUP.md.",
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
