package brain

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewBrainPaths_Layout(t *testing.T) {
	tmp := t.TempDir()
	p := NewBrainPaths(tmp)
	if p.Root != tmp {
		t.Errorf("Root = %q, want %q", p.Root, tmp)
	}
	if got, want := p.Dir, filepath.Join(tmp, ".mobiai", "brain"); got != want {
		t.Errorf("Dir = %q, want %q", got, want)
	}
	if got, want := p.ConfigFile, filepath.Join(tmp, ".mobiai", "brain", "config.json"); got != want {
		t.Errorf("ConfigFile = %q, want %q", got, want)
	}
	if got, want := p.ScanFile, filepath.Join(tmp, ".mobiai", "brain", "scan.json"); got != want {
		t.Errorf("ScanFile = %q, want %q", got, want)
	}
	if got, want := p.MemoriesDir, filepath.Join(tmp, ".mobiai", "brain", "memories"); got != want {
		t.Errorf("MemoriesDir = %q, want %q", got, want)
	}
}

func TestEnsureDirs_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	p := NewBrainPaths(tmp)
	if err := p.EnsureDirs(); err != nil {
		t.Fatal(err)
	}
	// Second call must not error.
	if err := p.EnsureDirs(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p.MemoriesDir); err != nil {
		t.Errorf("memories dir not created: %v", err)
	}
}

func TestExists(t *testing.T) {
	tmp := t.TempDir()
	p := NewBrainPaths(tmp)
	if p.Exists() {
		t.Fatal("Exists should be false on a fresh dir")
	}
	if err := p.EnsureDirs(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p.ConfigFile, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !p.Exists() {
		t.Error("Exists should be true after writing config.json")
	}
}

func TestFindProjectRoot(t *testing.T) {
	type setup func(t *testing.T, root string)
	cases := []struct {
		name    string
		build   setup
		startIn string // relative path inside root where we start the search
		want    RootSource
	}{
		{
			name: "existing brain wins over git",
			build: func(t *testing.T, root string) {
				mustMkdir(t, filepath.Join(root, ".git"))
				mustMkdir(t, filepath.Join(root, ".mobiai", "brain"))
				mustWrite(t, filepath.Join(root, ".mobiai", "brain", "config.json"), "{}")
			},
			startIn: ".",
			want:    RootSourceBrain,
		},
		{
			name: "git marker",
			build: func(t *testing.T, root string) {
				mustMkdir(t, filepath.Join(root, ".git"))
			},
			startIn: ".",
			want:    RootSourceGit,
		},
		{
			name: "settings.gradle.kts (Android/KMP)",
			build: func(t *testing.T, root string) {
				mustWrite(t, filepath.Join(root, "settings.gradle.kts"), "")
			},
			startIn: ".",
			want:    RootSourceGradle,
		},
		{
			name: "pubspec.yaml (Flutter)",
			build: func(t *testing.T, root string) {
				mustWrite(t, filepath.Join(root, "pubspec.yaml"), "name: x\n")
			},
			startIn: ".",
			want:    RootSourcePubspec,
		},
		{
			name: "Package.swift (iOS SPM)",
			build: func(t *testing.T, root string) {
				mustWrite(t, filepath.Join(root, "Package.swift"), "")
			},
			startIn: ".",
			want:    RootSourceSwift,
		},
		{
			name: "xcworkspace dir (iOS Xcode)",
			build: func(t *testing.T, root string) {
				mustMkdir(t, filepath.Join(root, "App.xcworkspace"))
			},
			startIn: ".",
			want:    RootSourceXcode,
		},
		{
			name: "Podfile only",
			build: func(t *testing.T, root string) {
				mustWrite(t, filepath.Join(root, "Podfile"), "")
			},
			startIn: ".",
			want:    RootSourcePodfile,
		},
		{
			name: "package.json with react-native",
			build: func(t *testing.T, root string) {
				mustWrite(t, filepath.Join(root, "package.json"), `{"dependencies": {"react-native": "0.74.0"}}`)
			},
			startIn: ".",
			want:    RootSourceRN,
		},
		{
			name: "package.json without react-native is not enough",
			build: func(t *testing.T, root string) {
				mustWrite(t, filepath.Join(root, "package.json"), `{"dependencies": {"react": "18"}}`)
			},
			startIn: ".",
			want:    RootSourceCwd,
		},
		{
			name: "walks up from nested dir",
			build: func(t *testing.T, root string) {
				mustMkdir(t, filepath.Join(root, ".git"))
				mustMkdir(t, filepath.Join(root, "a", "b", "c"))
			},
			startIn: filepath.Join("a", "b", "c"),
			want:    RootSourceGit,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			tc.build(t, root)
			start := filepath.Join(root, tc.startIn)
			info, err := FindProjectRoot(start)
			if err != nil {
				t.Fatal(err)
			}
			if info.Source != tc.want {
				t.Errorf("Source = %q, want %q (path=%q warning=%q)",
					info.Source, tc.want, info.Path, info.Warning)
			}
			if tc.want == RootSourceCwd && info.Warning == "" {
				t.Error("expected warning when falling back to cwd")
			}
		})
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
