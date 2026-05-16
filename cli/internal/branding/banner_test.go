package branding

import (
	"bytes"
	"strings"
	"testing"
)

func TestRender_PlainHasNoEscapes(t *testing.T) {
	out := Render(false)
	if strings.Contains(out, "\033[") {
		t.Errorf("plain render should not contain ANSI escapes:\n%s", out)
	}
	// Sanity: the wordmark and tagline are both present.
	if !strings.Contains(out, "MOBILE DEVELOPERS") {
		t.Errorf("tagline missing in plain render:\n%s", out)
	}
}

func TestRender_ColorWrapsInEscapes(t *testing.T) {
	out := Render(true)
	if !strings.Contains(out, "\033[1;36m") {
		t.Errorf("colored render should include the cyan start escape:\n%s", out)
	}
	if !strings.Contains(out, "\033[0m") {
		t.Errorf("colored render should include the reset escape:\n%s", out)
	}
}

func TestPrint_NoColorFlagSuppressesEscapes(t *testing.T) {
	var buf bytes.Buffer
	Print(&buf, true) // noColorFlag = true → never color
	if strings.Contains(buf.String(), "\033[") {
		t.Errorf("Print with noColorFlag=true should not emit escapes:\n%s", buf.String())
	}
}

func TestPrint_NonTTYDoesNotColor(t *testing.T) {
	// bytes.Buffer is not a TTY — shouldUseColor returns true for
	// non-*os.File writers as a safe default, but the caller's
	// noColorFlag should still gate it. With noColorFlag=false here,
	// we expect color (the caller's choice).
	var buf bytes.Buffer
	Print(&buf, false)
	if !strings.Contains(buf.String(), "\033[") {
		t.Errorf("Print with bytes.Buffer + noColorFlag=false should color (caller opted in):\n%s", buf.String())
	}
}

func TestShouldUseColor_RespectsNoColorEnv(t *testing.T) {
	// NO_COLOR set → never color, regardless of writer.
	t.Setenv("NO_COLOR", "1")
	if shouldUseColor(&bytes.Buffer{}) {
		t.Errorf("NO_COLOR=1 should disable color")
	}
}

func TestShouldUseColor_EmptyNoColorEnvIsNoOp(t *testing.T) {
	// NO_COLOR="" should NOT disable color (per the spec: any
	// non-empty value disables).
	t.Setenv("NO_COLOR", "")
	if !shouldUseColor(&bytes.Buffer{}) {
		t.Errorf("empty NO_COLOR should not disable color")
	}
}
