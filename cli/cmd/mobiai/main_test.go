package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootVersionFlag(t *testing.T) {
	cmd := newRootCmd("0.1.0")
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "0.1.0") {
		t.Errorf("expected version 0.1.0 in output, got: %q", got)
	}
}

func TestRootShowsHelpWhenNoArgs(t *testing.T) {
	cmd := newRootCmd("0.1.0")
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "mobiai") {
		t.Errorf("expected 'mobiai' in help output, got: %q", got)
	}
}
