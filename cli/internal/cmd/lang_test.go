package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ArisGuimera/MobiAI-Core/cli/internal/i18n"
	"github.com/ArisGuimera/MobiAI-Core/cli/internal/state"
)

func TestLang_ShowDefault(t *testing.T) {
	t.Cleanup(func() { i18n.SetLang(i18n.EN) })
	i18n.SetLang(i18n.EN)

	cmd := NewLangCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Current language: en") {
		t.Errorf("expected current language en; got %q", out.String())
	}
}

func TestLang_SetPersistsAndConfirmsInNewLang(t *testing.T) {
	t.Cleanup(func() { i18n.SetLang(i18n.EN) })
	tmp := t.TempDir()
	t.Setenv("MOBIAI_HOME", filepath.Join(tmp, ".mobiai"))
	i18n.SetLang(i18n.EN)

	cmd := NewLangCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"es"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	// The confirmation prints in the NEW language (lang switched before print).
	if !strings.Contains(out.String(), "Idioma cambiado a es") {
		t.Errorf("confirmation should be in es; got %q", out.String())
	}

	// The choice persisted to config.json.
	p, _ := state.NewPaths()
	cfg, err := state.LoadConfig(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Lang != "es" {
		t.Errorf("persisted Lang: got %q, want es", cfg.Lang)
	}
}

func TestLang_InvalidRejected(t *testing.T) {
	t.Cleanup(func() { i18n.SetLang(i18n.EN) })

	cmd := NewLangCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"fr"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected an error for an unsupported language")
	}
}
