package brain

import "strings"

// techMatcher maps a label to a bucket on Scan, with the needles that
// trigger the match. Needles are matched case-insensitively against the
// content of every interesting file in the index. Order doesn't matter;
// dedupAllBuckets normalizes the final lists.
type techMatcher struct {
	Label   string
	Bucket  string // "ui" | "di" | "network" | "persistence" | "serialization" | "testing" | "integrations" | "build_systems"
	Needles []string
}

var techMatchers = []techMatcher{
	// UI
	{"compose_multiplatform", "ui", []string{"org.jetbrains.compose"}},
	{"compose", "ui", []string{"androidx.compose"}},
	{"swiftui", "ui", []string{"import SwiftUI"}},
	{"uikit", "ui", []string{"import UIKit"}},
	{"coil", "ui", []string{"io.coil-kt"}},
	{"glide", "ui", []string{"com.github.bumptech.glide"}},

	// DI
	{"koin", "di", []string{"io.insert-koin", "org.koin"}},
	{"hilt", "di", []string{"com.google.dagger.hilt", "dagger.hilt.android"}},
	{"dagger", "di", []string{"com.google.dagger:dagger"}},

	// Network
	{"ktor", "network", []string{"io.ktor"}},
	{"retrofit", "network", []string{"com.squareup.retrofit2"}},
	{"okhttp", "network", []string{"com.squareup.okhttp3"}},

	// Persistence
	{"room", "persistence", []string{"androidx.room"}},
	{"datastore", "persistence", []string{"androidx.datastore"}},
	{"sqldelight", "persistence", []string{"app.cash.sqldelight"}},
	{"realm", "persistence", []string{"io.realm"}},

	// Serialization
	{"kotlinx.serialization", "serialization", []string{"kotlinx.serialization", "plugin.serialization"}},
	{"gson", "serialization", []string{"com.google.code.gson"}},
	{"moshi", "serialization", []string{"com.squareup.moshi"}},

	// Testing
	{"junit", "testing", []string{"junit:junit", "org.junit"}},
	{"mockk", "testing", []string{"io.mockk"}},
	{"mockito", "testing", []string{"org.mockito"}},
	{"turbine", "testing", []string{"app.cash.turbine"}},
	{"espresso", "testing", []string{"androidx.test.espresso"}},

	// Integrations
	{"firebase", "integrations", []string{"com.google.firebase", "google-services", "Firebase/Core"}},
}

// detectDeps fans out the techMatchers across the index and bucketizes hits.
// Some signals (CI/CD, sensitive files) come from path-level checks rather
// than content matching — those live in dedicated helpers below.
func detectDeps(idx *scanIndex, s *Scan) {
	for _, m := range techMatchers {
		for _, needle := range m.Needles {
			if !idx.containsAny(needle) {
				continue
			}
			appendBucket(s, m.Bucket, m.Label)
			break
		}
	}
	detectCICD(idx, s)
	detectSensitiveFiles(idx, s)
}

func appendBucket(s *Scan, bucket, label string) {
	switch bucket {
	case "ui":
		s.UI = append(s.UI, label)
	case "di":
		s.DI = append(s.DI, label)
	case "network":
		s.Network = append(s.Network, label)
	case "persistence":
		s.Persistence = append(s.Persistence, label)
	case "serialization":
		s.Serialization = append(s.Serialization, label)
	case "testing":
		s.Testing = append(s.Testing, label)
	case "integrations":
		s.Integrations = append(s.Integrations, label)
	case "build_systems":
		s.BuildSystems = append(s.BuildSystems, label)
	}
}

// detectCICD looks for well-known CI files/dirs by path, not content,
// so we don't have to load every YAML in the repo into memory.
func detectCICD(idx *scanIndex, s *Scan) {
	for path, isDir := range idx.allPaths {
		switch {
		case isDir && (path == ".github/workflows" || strings.HasSuffix(path, "/.github/workflows")):
			s.CICD = append(s.CICD, "github_actions")
		case !isDir && (path == "bitrise.yml" || strings.HasSuffix(path, "/bitrise.yml")):
			s.CICD = append(s.CICD, "bitrise")
		case !isDir && (path == "codemagic.yaml" || strings.HasSuffix(path, "/codemagic.yaml")):
			s.CICD = append(s.CICD, "codemagic")
		case !isDir && (path == "fastlane/Fastfile" || strings.HasSuffix(path, "/fastlane/Fastfile")):
			s.CICD = append(s.CICD, "fastlane")
		}
	}
}

// detectSensitiveFiles records the existence of files that often carry
// secrets, but never reads their content. The user (and downstream
// agents) get a heads-up via warnings without leaking bytes anywhere.
func detectSensitiveFiles(idx *scanIndex, s *Scan) {
	for path, isDir := range idx.allPaths {
		if isDir {
			continue
		}
		switch {
		case path == ".env" || strings.HasSuffix(path, "/.env"):
			s.Warnings = append(s.Warnings, "sensitive file detected: "+path+" (not read)")
		case strings.HasSuffix(path, "/GoogleService-Info.plist") || path == "GoogleService-Info.plist":
			s.Warnings = append(s.Warnings, "sensitive file detected: "+path+" (not read)")
		case strings.HasSuffix(path, "/google-services.json") || path == "google-services.json":
			s.Warnings = append(s.Warnings, "sensitive file detected: "+path+" (not read)")
		}
	}
}
