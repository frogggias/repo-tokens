package detect

import "strings"

// Preset defines a recognized project type with its file patterns.
type Preset struct {
	Name       string
	Markers    []string // files/dirs whose presence indicates this type
	Extensions []string // file extensions to count (with dot)
	SkipDirs   []string // directory names to skip within this project
}

// Project is a project found during directory scanning.
type Project struct {
	Name   string
	Path   string  // relative path from scan root
	Preset *Preset

	// Set by WithoutTests:
	SkipTestDirSuffix bool     // skip dirs ending in Tests/Test/Specs
	TestFileSuffixes  []string // file suffixes like _test.go, .spec.ts
}

// Presets is the full list of marker-based language presets.
var Presets = []Preset{
	// Systems
	{Name: "go", Markers: []string{"go.mod"}, Extensions: []string{".go"}, SkipDirs: []string{"vendor", "testdata"}},
	{Name: "rust", Markers: []string{"Cargo.toml"}, Extensions: []string{".rs"}, SkipDirs: []string{"target"}},
	{Name: "cpp", Markers: []string{"CMakeLists.txt", "meson.build", "configure.ac", "Kconfig"}, Extensions: []string{".c", ".h", ".cpp", ".hpp", ".cc", ".cxx", ".hh"}, SkipDirs: []string{"build", "cmake-build-debug", "cmake-build-release"}},
	{Name: "zig", Markers: []string{"build.zig"}, Extensions: []string{".zig"}, SkipDirs: []string{"zig-cache", "zig-out"}},
	{Name: "nim", Markers: []string{"*.nimble"}, Extensions: []string{".nim", ".nims"}, SkipDirs: []string{"nimcache"}},
	{Name: "v", Markers: []string{"v.mod"}, Extensions: []string{".v"}},

	// Web
	{Name: "web", Markers: []string{"package.json"}, Extensions: []string{".js", ".ts", ".jsx", ".tsx", ".mjs", ".mts", ".html", ".css", ".scss", ".sass", ".less", ".vue", ".svelte", ".njk", ".liquid", ".astro", ".mdx"}, SkipDirs: []string{"dist", "build", ".next", ".nuxt", ".output", "_site", ".cache", ".parcel-cache", "coverage", "storybook-static", ".turbo"}},

	// Mobile
	{Name: "swift", Markers: []string{"Package.swift", "*.xcodeproj", "*.xcworkspace"}, Extensions: []string{".swift"}, SkipDirs: []string{"Pods", ".build", "DerivedData", "Derived", "build", ".swiftpm", "checkouts", "SourcePackages"}},
	{Name: "kotlin", Markers: []string{"build.gradle.kts"}, Extensions: []string{".kt", ".kts"}, SkipDirs: []string{"build", ".gradle", ".idea"}},
	{Name: "dart", Markers: []string{"pubspec.yaml"}, Extensions: []string{".dart"}, SkipDirs: []string{".dart_tool", "build"}},

	// JVM
	{Name: "java", Markers: []string{"build.gradle", "pom.xml"}, Extensions: []string{".java"}, SkipDirs: []string{"build", "target", ".gradle", ".idea"}},
	{Name: "scala", Markers: []string{"build.sbt"}, Extensions: []string{".scala", ".sc"}, SkipDirs: []string{"target", "project/target", ".bsp"}},
	{Name: "clojure", Markers: []string{"project.clj", "deps.edn"}, Extensions: []string{".clj", ".cljs", ".cljc", ".edn"}, SkipDirs: []string{"target", ".cpcache"}},

	// Scripting
	{Name: "python", Markers: []string{"pyproject.toml", "setup.py", "setup.cfg", "requirements.txt"}, Extensions: []string{".py", ".pyi"}, SkipDirs: []string{"venv", ".venv", "__pycache__", ".tox", ".mypy_cache", "site-packages", ".eggs", ".pytest_cache", "htmlcov"}},
	{Name: "ruby", Markers: []string{"Gemfile"}, Extensions: []string{".rb", ".erb", ".rake"}, SkipDirs: []string{"vendor"}},
	{Name: "php", Markers: []string{"composer.json"}, Extensions: []string{".php"}, SkipDirs: []string{"vendor"}},
	{Name: "perl", Markers: []string{"Makefile.PL", "cpanfile", "dist.ini"}, Extensions: []string{".pl", ".pm", ".t"}, SkipDirs: []string{"blib"}},
	{Name: "lua", Markers: []string{"*.rockspec", ".luacheckrc"}, Extensions: []string{".lua"}, SkipDirs: []string{"lua_modules"}},

	// .NET
	{Name: "csharp", Markers: []string{"*.csproj", "*.sln"}, Extensions: []string{".cs"}, SkipDirs: []string{"bin", "obj", ".vs"}},
	{Name: "fsharp", Markers: []string{"*.fsproj"}, Extensions: []string{".fs", ".fsi", ".fsx"}, SkipDirs: []string{"bin", "obj", ".vs"}},

	// Functional
	{Name: "elixir", Markers: []string{"mix.exs"}, Extensions: []string{".ex", ".exs", ".heex", ".leex"}, SkipDirs: []string{"_build", "deps"}},
	{Name: "haskell", Markers: []string{"*.cabal", "stack.yaml"}, Extensions: []string{".hs", ".lhs"}, SkipDirs: []string{".stack-work", "dist-newstyle"}},
	{Name: "ocaml", Markers: []string{"dune-project", "*.opam"}, Extensions: []string{".ml", ".mli"}, SkipDirs: []string{"_build", "_opam"}},
	{Name: "gleam", Markers: []string{"gleam.toml"}, Extensions: []string{".gleam"}, SkipDirs: []string{"build"}},
	{Name: "erlang", Markers: []string{"rebar.config"}, Extensions: []string{".erl", ".hrl"}, SkipDirs: []string{"_build", "_rel"}},

	// Scientific
	{Name: "julia", Markers: []string{"Project.toml"}, Extensions: []string{".jl"}},
	{Name: "r", Markers: []string{"DESCRIPTION", ".Rproj"}, Extensions: []string{".R", ".r", ".Rmd"}, SkipDirs: []string{"renv"}},

	// Infrastructure
	{Name: "terraform", Markers: []string{"*.tf"}, Extensions: []string{".tf", ".tfvars"}, SkipDirs: []string{".terraform"}},
}

// ExtensionOnlyPresets are identified purely by file extension (no markers).
var ExtensionOnlyPresets = []Preset{
	{Name: "shell", Extensions: []string{".sh", ".bash", ".zsh", ".fish"}},
	{Name: "sql", Extensions: []string{".sql"}},
	{Name: "graphql", Extensions: []string{".graphql", ".gql"}},
	{Name: "protobuf", Extensions: []string{".proto"}},
	{Name: "yaml", Extensions: []string{".yaml", ".yml"}},
	{Name: "toml", Extensions: []string{".toml"}},
	{Name: "markdown", Extensions: []string{".md"}},
	{Name: "latex", Extensions: []string{".tex", ".bib", ".sty", ".cls"}},
}

// GlobalSkipDirs are always skipped regardless of project type.
var GlobalSkipDirs = map[string]bool{
	".git": true, ".svn": true, ".hg": true,
	"node_modules": true, "__pycache__": true,
	".idea": true, ".vscode": true,
	".github": true, ".gitlab": true,
}

// GlobalSkipFiles are always skipped.
var GlobalSkipFiles = map[string]bool{
	".DS_Store": true,
}

var testDirSuffixes = []string{"tests", "test", "specs", "spec", "uitests"}
var testDirExact = []string{"test", "tests", "spec", "specs", "__tests__", "test_helpers", "testdata", "fixtures"}
var testFileSuffixes = []string{
	"_test.go",
	".test.ts", ".test.tsx", ".test.js", ".test.jsx", ".test.mjs",
	".spec.ts", ".spec.tsx", ".spec.js", ".spec.jsx", ".spec.mjs",
	"_test.py", "_spec.rb", "_test.rs", "_test.dart", "_test.exs",
}

// WithoutTests returns a copy of the project with test directories and files excluded.
func WithoutTests(p Project) Project {
	// Copy SkipDirs to avoid mutating the preset's backing array.
	skipDirs := make([]string, 0, len(p.Preset.SkipDirs)+len(testDirExact))
	skipDirs = append(skipDirs, p.Preset.SkipDirs...)
	skipDirs = append(skipDirs, testDirExact...)

	preset := *p.Preset
	preset.SkipDirs = skipDirs

	p.Preset = &preset
	p.SkipTestDirSuffix = true
	p.TestFileSuffixes = testFileSuffixes
	return p
}

// IsTestDir returns true if a dir name looks like a test directory
// (e.g. "VaultTests", "MyAppUITests").
func IsTestDir(name string) bool {
	lower := strings.ToLower(name)
	for _, s := range testDirSuffixes {
		if strings.HasSuffix(lower, s) && lower != s {
			return true
		}
	}
	return false
}

// MergeSkipDirs combines GlobalSkipDirs with preset-specific dirs into a lookup set.
func MergeSkipDirs(extra []string) map[string]bool {
	m := make(map[string]bool, len(GlobalSkipDirs)+len(extra))
	for k := range GlobalSkipDirs {
		m[k] = true
	}
	for _, d := range extra {
		m[d] = true
	}
	return m
}

// ExtensionSet converts a slice of extensions into a lookup set.
func ExtensionSet(exts []string) map[string]bool {
	m := make(map[string]bool, len(exts))
	for _, e := range exts {
		m[e] = true
	}
	return m
}

// IsMarkerPreset returns true if the named preset uses project markers
// (as opposed to extension-only presets like shell, yaml, markdown).
func IsMarkerPreset(name string) bool {
	for _, p := range Presets {
		if p.Name == name {
			return true
		}
	}
	return false
}

// FindPreset returns the preset with the given name, or nil.
func FindPreset(name string) *Preset {
	for i := range Presets {
		if Presets[i].Name == name {
			return &Presets[i]
		}
	}
	for i := range ExtensionOnlyPresets {
		if ExtensionOnlyPresets[i].Name == name {
			return &ExtensionOnlyPresets[i]
		}
	}
	return nil
}

// GenericPreset returns a preset that includes all known source extensions.
func GenericPreset() Preset {
	seen := make(map[string]bool)
	var exts, dirs []string
	for _, list := range [][]Preset{Presets, ExtensionOnlyPresets} {
		for _, p := range list {
			for _, ext := range p.Extensions {
				if !seen[ext] {
					seen[ext] = true
					exts = append(exts, ext)
				}
			}
		}
	}
	seenD := make(map[string]bool)
	for _, p := range Presets {
		for _, d := range p.SkipDirs {
			if !seenD[d] {
				seenD[d] = true
				dirs = append(dirs, d)
			}
		}
	}
	return Preset{Name: "generic", Extensions: exts, SkipDirs: dirs}
}

// BuildExtToLang builds a reverse lookup from file extension to language name.
func BuildExtToLang() map[string]string {
	m := make(map[string]string)
	for _, list := range [][]Preset{Presets, ExtensionOnlyPresets} {
		for _, p := range list {
			for _, ext := range p.Extensions {
				if _, ok := m[ext]; !ok {
					m[ext] = p.Name
				}
			}
		}
	}
	return m
}
