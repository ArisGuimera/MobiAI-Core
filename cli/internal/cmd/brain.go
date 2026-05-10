package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/brain"
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
	root.AddCommand(newBrainInitCmd(), newBrainScanCmd(), newBrainContextCmd())
	return root
}

// brainCommonFlags wires up shared --root/--cwd flags on a leaf command.
// Both flags resolve to the same value (the project root); having two
// names lets users pick whichever reads better in their shell history.
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
		RunE:  runBrainContext,
	}
	brainCommonFlags(c)
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
	md := brain.BuildContext(cfg, scan, paths)
	fmt.Fprint(c.OutOrStdout(), md)
	return nil
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
