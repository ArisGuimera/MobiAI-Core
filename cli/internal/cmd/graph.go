package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/graph"
)

// NewGraphCmd builds `mobiai graph <init|status|search|callers|context>`.
//
// Phase 1 of MobiAI Graph — a per-project semantic code index for mobile
// codebases. State lives at <root>/.mobiai/graph/, never in the global
// ~/.mobiai/ directory (that's the CLI's own state).
func NewGraphCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "graph",
		Short: "Exploración semántica del código mobile",
		Long: "MobiAI Graph indexa el proyecto y permite buscar símbolos, encontrar referencias " +
			"y obtener contexto relevante para una tarea. Vive en <repo>/.mobiai/graph/.",
	}
	root.AddCommand(
		newGraphInitCmd(),
		newGraphStatusCmd(),
		newGraphSearchCmd(),
		newGraphCallersCmd(),
		newGraphContextCmd(),
	)
	return root
}

// graphCommonFlags wires up the shared --root flag on a leaf command.
// When omitted, the command resolves the project root from cwd.
func graphCommonFlags(c *cobra.Command) {
	c.Flags().String("root", "", "ruta del proyecto (default: cwd)")
}

// resolveGraphRoot returns the absolute project root for a graph command.
// Honors --root if given; otherwise falls back to the current working
// directory. Unlike Brain, Graph does NOT walk up looking for project
// markers — it operates on whatever directory the user points it at.
func resolveGraphRoot(c *cobra.Command) (string, error) {
	if explicit, _ := c.Flags().GetString("root"); explicit != "" {
		abs, err := filepath.Abs(explicit)
		if err != nil {
			return "", fmt.Errorf("absolutizar --root: %w", err)
		}
		info, err := os.Stat(abs)
		if err != nil || !info.IsDir() {
			return "", fmt.Errorf("--root no es un directorio: %s", abs)
		}
		return abs, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	return filepath.Abs(cwd)
}

// graphIndexPath returns the canonical on-disk location of the Graph
// index for the project at root.
func graphIndexPath(root string) string {
	return filepath.Join(root, ".mobiai", "graph", "index.json")
}

// loadGraphIndex reads the index for root, returning a friendly Spanish
// error suggesting `mobiai graph init` when the file is missing.
func loadGraphIndex(root string) (*graph.Index, string, error) {
	path := graphIndexPath(root)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, path, fmt.Errorf("no encontré %s. Corré `mobiai graph init` primero", path)
	}
	idx, err := graph.Read(path)
	if err != nil {
		return nil, path, fmt.Errorf("leer índice %s: %w", path, err)
	}
	return idx, path, nil
}

func newGraphInitCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "init",
		Short: "Indexa el proyecto y guarda .mobiai/graph/index.json",
		RunE:  runGraphInit,
	}
	graphCommonFlags(c)
	return c
}

func runGraphInit(c *cobra.Command, _ []string) error {
	out := c.OutOrStdout()
	root, err := resolveGraphRoot(c)
	if err != nil {
		return err
	}

	idx, err := graph.Build(root)
	if err != nil {
		return fmt.Errorf("indexar proyecto: %w", err)
	}

	path := graphIndexPath(root)
	if err := graph.Write(path, idx); err != nil {
		return fmt.Errorf("escribir índice: %w", err)
	}

	kotlin, swift := countByLanguage(idx)
	symbols := countSymbols(idx)

	fmt.Fprintf(out, "✓ Índice generado: %s\n", path)
	fmt.Fprintf(out, "  Archivos: %d\n", len(idx.Files))
	fmt.Fprintf(out, "  Símbolos: %d\n", symbols)
	fmt.Fprintf(out, "  Kotlin: %d\n", kotlin)
	fmt.Fprintf(out, "  Swift: %d\n", swift)
	return nil
}

func newGraphStatusCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "status",
		Short: "Muestra info del índice actual",
		RunE:  runGraphStatus,
	}
	graphCommonFlags(c)
	return c
}

func runGraphStatus(c *cobra.Command, _ []string) error {
	out := c.OutOrStdout()
	root, err := resolveGraphRoot(c)
	if err != nil {
		return err
	}
	idx, path, err := loadGraphIndex(root)
	if err != nil {
		return err
	}

	kotlin, swift := countByLanguage(idx)
	symbols := countSymbols(idx)

	fmt.Fprintf(out, "Índice: %s\n", path)
	fmt.Fprintf(out, "Generado: %s (%s)\n",
		idx.GeneratedAt.Format(time.RFC3339), humanizeAge(idx.GeneratedAt))
	fmt.Fprintf(out, "Archivos: %d\n", len(idx.Files))
	fmt.Fprintf(out, "Símbolos: %d\n", symbols)
	fmt.Fprintf(out, "Kotlin: %d\n", kotlin)
	fmt.Fprintf(out, "Swift: %d\n", swift)
	return nil
}

func newGraphSearchCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "search <término>",
		Short: "Busca símbolos por nombre",
		Args:  cobra.ExactArgs(1),
		RunE:  runGraphSearch,
	}
	graphCommonFlags(c)
	c.Flags().Int("limit", 0, "máximo de resultados a mostrar (0 = sin límite)")
	c.Flags().String("kind", "", "filtrar por tipo de símbolo (class, fun, object, interface, struct, enum, protocol, actor, func, extension)")
	return c
}

func runGraphSearch(c *cobra.Command, args []string) error {
	out := c.OutOrStdout()
	root, err := resolveGraphRoot(c)
	if err != nil {
		return err
	}
	idx, _, err := loadGraphIndex(root)
	if err != nil {
		return err
	}
	limit, _ := c.Flags().GetInt("limit")
	kind, _ := c.Flags().GetString("kind")

	hits := graph.Search(idx, args[0])

	// Filter by kind (case-insensitive) when --kind is provided. Done in the
	// cmd layer instead of pushing it into graph.Search so the package API
	// stays minimal — this is a UI concern.
	if kind != "" {
		kindLower := strings.ToLower(kind)
		filtered := hits[:0:0]
		for _, h := range hits {
			if strings.ToLower(h.Symbol.Kind) == kindLower {
				filtered = append(filtered, h)
			}
		}
		hits = filtered
	}

	if len(hits) == 0 {
		fmt.Fprintln(out, "Sin coincidencias.")
		return nil
	}

	truncated := false
	if limit > 0 && len(hits) > limit {
		hits = hits[:limit]
		truncated = true
	}

	for _, h := range hits {
		fmt.Fprintf(out, "%s:%d %s %s\n", h.File, h.Symbol.Line, h.Symbol.Kind, h.Symbol.Name)
	}
	if truncated {
		fmt.Fprintf(out, "… (mostrando %d, usá --limit 0 para ver todos)\n", limit)
	}
	return nil
}

func newGraphCallersCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "callers <símbolo>",
		Short: "Lista referencias textuales del símbolo",
		Args:  cobra.ExactArgs(1),
		RunE:  runGraphCallers,
	}
	graphCommonFlags(c)
	return c
}

func runGraphCallers(c *cobra.Command, args []string) error {
	out := c.OutOrStdout()
	root, err := resolveGraphRoot(c)
	if err != nil {
		return err
	}
	idx, _, err := loadGraphIndex(root)
	if err != nil {
		return err
	}
	hits := graph.Callers(idx, root, args[0])
	if len(hits) == 0 {
		fmt.Fprintln(out, "Sin referencias.")
		return nil
	}
	for _, h := range hits {
		fmt.Fprintf(out, "%s:%d %s\n", h.File, h.Line, h.Snippet)
	}
	return nil
}

func newGraphContextCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "context <tarea...>",
		Short: "Archivos relevantes para una tarea (heurística)",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runGraphContext,
	}
	graphCommonFlags(c)
	return c
}

func runGraphContext(c *cobra.Command, args []string) error {
	out := c.OutOrStdout()
	root, err := resolveGraphRoot(c)
	if err != nil {
		return err
	}
	idx, _, err := loadGraphIndex(root)
	if err != nil {
		return err
	}
	task := joinArgs(args)
	hits := graph.Context(idx, task, 10)
	if len(hits) == 0 {
		fmt.Fprintln(out, "Sin archivos relevantes encontrados.")
		return nil
	}
	for _, h := range hits {
		fmt.Fprintf(out, "%s  (score %d)\n", h.File, h.Score)
	}
	return nil
}

// joinArgs joins free-form task tokens with single spaces. Done as a
// helper (instead of strings.Join inline) so it's easy to swap in
// smarter normalization later without touching the command body.
func joinArgs(args []string) string {
	out := ""
	for i, a := range args {
		if i > 0 {
			out += " "
		}
		out += a
	}
	return out
}

// countByLanguage returns (kotlin, swift) file counts in idx.
func countByLanguage(idx *graph.Index) (int, int) {
	var k, s int
	for _, f := range idx.Files {
		switch f.Language {
		case "kotlin":
			k++
		case "swift":
			s++
		}
	}
	return k, s
}

// countSymbols sums the total number of symbols across all files.
func countSymbols(idx *graph.Index) int {
	n := 0
	for _, f := range idx.Files {
		n += len(f.Symbols)
	}
	return n
}

// humanizeAge renders a Spanish "hace Xm/Xh/Xd" relative timestamp for
// a past time t. Falls back to "hace Xm" with X=0 for very fresh times.
func humanizeAge(t time.Time) string {
	d := time.Since(t)
	if d < 0 {
		d = 0
	}
	if d < time.Hour {
		mins := int(d.Round(time.Minute) / time.Minute)
		return fmt.Sprintf("hace %dm", mins)
	}
	if d < 24*time.Hour {
		hours := int(d.Round(time.Hour) / time.Hour)
		return fmt.Sprintf("hace %dh", hours)
	}
	days := int(d.Round(24*time.Hour) / (24 * time.Hour))
	return fmt.Sprintf("hace %dd", days)
}
