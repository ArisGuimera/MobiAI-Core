package brain

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const sampleCard = `## Login crash on screen rotation

- id: login-crash-on-screen-rotation-20260914-163044
- type: bug_fix
- status: active
- platform: android
- area: auth
- date: 2026-09-14
- verified: 2026-09-14

The form state lived in the Activity; moved it to the ViewModel.

### Files
- app/src/main/java/com/example/auth/LoginViewModel.kt
`

func TestParseCard_TitleMetaBody(t *testing.T) {
	e := ParseCardContent(sampleCard)
	if e.Title != "Login crash on screen rotation" {
		t.Errorf("Title = %q", e.Title)
	}
	for k, want := range map[string]string{
		"id":       "login-crash-on-screen-rotation-20260914-163044",
		"type":     "bug_fix",
		"status":   "active",
		"platform": "android",
		"area":     "auth",
		"date":     "2026-09-14",
		"verified": "2026-09-14",
	} {
		if got := e.Get(k); got != want {
			t.Errorf("meta %s = %q, want %q", k, got, want)
		}
	}
	if !strings.HasPrefix(e.Body, "The form state lived in the Activity") {
		t.Errorf("Body should start with the prose, got:\n%s", e.Body)
	}
	if !strings.Contains(e.Body, "### Files") {
		t.Errorf("Body should keep the ### Files section, got:\n%s", e.Body)
	}
}

func TestParseCard_H2InsideBodyStaysInBody(t *testing.T) {
	content := `## Login crash

- id: login-crash-20260928-185828
- type: bug_fix
- status: active
- date: 2026-09-28

## Symptom
The app crashes when the device rotates.

## Root cause
The form state is kept in the Activity.
`
	e := ParseCardContent(content)
	if e.Title != "Login crash" {
		t.Errorf("Title = %q, want the first ## line only", e.Title)
	}
	for _, want := range []string{"## Symptom", "The app crashes", "## Root cause", "kept in the Activity"} {
		if !strings.Contains(e.Body, want) {
			t.Errorf("Body lost %q; body:\n%s", want, e.Body)
		}
	}
}

func TestParseCard_CRLF(t *testing.T) {
	lf := ParseCardContent(sampleCard)
	crlf := ParseCardContent(strings.ReplaceAll(sampleCard, "\n", "\r\n"))
	if !reflect.DeepEqual(lf, crlf) {
		t.Errorf("CRLF card parsed differently:\nLF:   %#v\nCRLF: %#v", lf, crlf)
	}
}

func TestParseCard_FilesSection(t *testing.T) {
	content := `## Use Koin

- id: use-koin-20260510-123045
- date: 2026-05-10

### Decision
Use Koin.

### Files
- composeApp/src/commonMain/Module.kt
- ` + "`composeApp/build.gradle.kts`" + `

### Notes
- not a file, a regular bullet
`
	e := ParseCardContent(content)
	want := []string{"composeApp/src/commonMain/Module.kt", "composeApp/build.gradle.kts"}
	if !reflect.DeepEqual(e.Files, want) {
		t.Errorf("Files = %#v, want %#v", e.Files, want)
	}
}

func TestParseCard_FileWithoutIDUsesFilename(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "hand-written-note-20260101-000000.md")
	mustWrite(t, path, "## Hand written\n\n- status: active\n\nBody.\n")
	e, err := ParseCard(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := e.Get("id"); got != "hand-written-note-20260101-000000" {
		t.Errorf("id = %q, want the filename without .md", got)
	}
	if e.File != "hand-written-note-20260101-000000.md" {
		t.Errorf("File = %q", e.File)
	}
}
