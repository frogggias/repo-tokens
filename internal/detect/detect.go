package detect

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ByMarkers walks the directory tree looking for project markers.
// Stops recursing into directories that are already detected as projects.
func ByMarkers(root string) []Project {
	if preset := matchPreset(root); preset != nil {
		return []Project{{
			Name:   filepath.Base(root),
			Path:   ".",
			Preset: preset,
		}}
	}

	var projects []Project
	detected := make(map[string]bool)

	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if rel == "." {
			return nil
		}
		if depth := strings.Count(rel, string(filepath.Separator)) + 1; depth > 4 {
			return filepath.SkipDir
		}

		name := d.Name()
		if GlobalSkipDirs[name] || strings.HasPrefix(name, ".") {
			return filepath.SkipDir
		}
		for dp := range detected {
			if strings.HasPrefix(rel, dp+string(filepath.Separator)) {
				return filepath.SkipDir
			}
		}

		if preset := matchPreset(path); preset != nil {
			projects = append(projects, Project{Name: name, Path: rel, Preset: preset})
			detected[rel] = true
			return filepath.SkipDir
		}
		return nil
	})
	return projects
}

// ByExtensions scans all files and groups by language based on extension.
// Used as fallback when no project markers are found.
func ByExtensions(root string) []Project {
	extToLang := BuildExtToLang()
	counts := make(map[string]int)

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if GlobalSkipDirs[name] || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if GlobalSkipFiles[d.Name()] {
			return nil
		}
		if lang, ok := extToLang[strings.ToLower(filepath.Ext(path))]; ok {
			counts[lang]++
		}
		return nil
	})

	if len(counts) == 0 {
		return nil
	}

	type lc struct {
		lang  string
		count int
	}
	sorted := make([]lc, 0, len(counts))
	for lang, n := range counts {
		sorted = append(sorted, lc{lang, n})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].count > sorted[j].count })

	projects := make([]Project, 0, len(sorted))
	for _, s := range sorted {
		if p := FindPreset(s.lang); p != nil {
			projects = append(projects, Project{Name: s.lang, Path: ".", Preset: p})
		}
	}
	return projects
}

func matchPreset(dir string) *Preset {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	names := make(map[string]bool, len(entries))
	for _, e := range entries {
		names[e.Name()] = true
	}

	for i := range Presets {
		p := &Presets[i]
		for _, marker := range p.Markers {
			if strings.Contains(marker, "*") {
				for name := range names {
					if matched, _ := filepath.Match(marker, name); matched {
						return p
					}
				}
			} else if names[marker] {
				return p
			}
		}
	}
	return nil
}
