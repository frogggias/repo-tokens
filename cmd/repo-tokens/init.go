package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ehmo/repo-tokens/internal/detect"
	"github.com/ehmo/repo-tokens/internal/readme"
)

func runInit() {
	dir := "."
	if len(os.Args) > 2 {
		dir = os.Args[2]
	}
	dir, _ = filepath.Abs(dir)
	branch := defaultBranch(dir)

	projects := detect.ByMarkers(dir)
	if len(projects) == 0 {
		projects = detect.ByExtensions(dir)
	}

	fmt.Println("repo-tokens init")
	fmt.Println()
	if len(projects) > 0 {
		fmt.Printf("Detected %d project(s):\n", len(projects))
		for _, p := range projects {
			fmt.Printf("  %s (%s)\n", p.Path, p.Preset.Name)
		}
		fmt.Println()
	} else {
		fmt.Println("No specific project type detected — will use generic scanning.")
	}

	// Workflow
	wfPath := filepath.Join(dir, ".github", "workflows", "repo-tokens.yml")
	if _, err := os.Stat(wfPath); err == nil {
		fmt.Println("Workflow already exists:", rel(dir, wfPath))
	} else if err := writeFile(wfPath, workflow(branch)); err != nil {
		fmt.Fprintf(os.Stderr, "error: workflow: %v\n", err)
	} else {
		fmt.Println("Created", rel(dir, wfPath))
	}

	// README markers
	readmePath := filepath.Join(dir, "README.md")
	if _, err := os.Stat(readmePath); err == nil {
		if ok, err := readme.InsertMarkers(readmePath, "token-count"); err != nil {
			fmt.Fprintf(os.Stderr, "error: readme: %v\n", err)
		} else if ok {
			fmt.Println("Added token-count markers to README.md")
		} else {
			fmt.Println("README.md already has token-count markers")
		}
	} else {
		fmt.Println("No README.md found — skipping marker insertion")
	}

	// Config for monorepos
	if len(projects) > 1 {
		cfgPath := filepath.Join(dir, ".repo-tokens.yml")
		if _, err := os.Stat(cfgPath); err == nil {
			fmt.Println("Config already exists: .repo-tokens.yml")
		} else if err := os.WriteFile(cfgPath, []byte(configYAML(projects)), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "error: config: %v\n", err)
		} else {
			fmt.Println("Created .repo-tokens.yml (monorepo detected)")
		}
	}

	fmt.Printf("\nDone. Push to %s to trigger the first token count.\n", branch)
}

func workflow(branch string) string {
	return fmt.Sprintf(`name: Update token count

on:
  push:
    branches: [%s]

permissions:
  contents: write

jobs:
  tokens:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: ehmo/repo-tokens@v1
        id: tokens

      - name: Commit if changed
        run: |
          git add -A .github/badges/ README.md
          git diff --cached --quiet && exit 0
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git commit -m "docs: update token count to ${{ steps.tokens.outputs.badge }}"
          git push
`, branch)
}

func configYAML(projects []detect.Project) string {
	var b strings.Builder
	b.WriteString("# repo-tokens configuration\n# See https://github.com/ehmo/repo-tokens\n\ncontext-window: 200000\n\nprojects:\n")
	for _, p := range projects {
		fmt.Fprintf(&b, "  - name: %s\n    path: %s\n    type: %s\n", p.Name, p.Path, p.Preset.Name)
	}
	return b.String()
}

func defaultBranch(dir string) string {
	gitDir := filepath.Join(dir, ".git")
	for _, branch := range []string{"main", "master"} {
		if _, err := os.Stat(filepath.Join(gitDir, "refs", "heads", branch)); err == nil {
			return branch
		}
	}
	if head, err := os.ReadFile(filepath.Join(gitDir, "HEAD")); err == nil {
		s := strings.TrimSpace(string(head))
		if b, ok := strings.CutPrefix(s, "ref: refs/heads/"); ok {
			return b
		}
	}
	return "main"
}

func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func rel(base, path string) string {
	if r, err := filepath.Rel(base, path); err == nil {
		return r
	}
	return path
}
