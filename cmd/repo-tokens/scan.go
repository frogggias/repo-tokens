package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ehmo/repo-tokens/internal/config"
	"github.com/ehmo/repo-tokens/internal/counter"
	"github.com/ehmo/repo-tokens/internal/detect"
)

// ScanParams configures a token counting scan.
type ScanParams struct {
	Target        string
	ContextWindow int
	Encoding      string
	IncludeTests  bool
}

// ScanResult holds the aggregated results of a scan.
type ScanResult struct {
	Results       []counter.Result
	Projects      []detect.Project
	TotalFiles    int
	TotalTokens   int
	ContextWindow int
	Encoding      string
}

// scan runs the full detect → count pipeline.
func scan(p ScanParams) ScanResult {
	cfg, err := config.Load(p.Target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: config: %v\n", err)
	}

	ctxWin := p.ContextWindow
	encoding := p.Encoding
	if cfg != nil {
		if cfg.ContextWindow != 0 {
			ctxWin = cfg.ContextWindow
		}
		if cfg.Encoding != "" {
			encoding = cfg.Encoding
		}
	}

	// Detection chain: config → markers → extensions → generic fallback
	var projects []detect.Project
	if cfg != nil && len(cfg.Projects) > 0 {
		projects = config.ToProjects(cfg)
	} else {
		projects = detect.ByMarkers(p.Target)
		if len(projects) == 0 {
			projects = detect.ByExtensions(p.Target)
		} else {
			// Supplement with extension-based detection for any
			// marker-based language not already covered.
			// This catches cases like C++ repos with Python build systems.
			covered := make(map[string]bool)
			for _, proj := range projects {
				covered[proj.Preset.Name] = true
			}
			for _, ext := range detect.ByExtensions(p.Target) {
				if covered[ext.Preset.Name] || !detect.IsMarkerPreset(ext.Preset.Name) {
					continue
				}
				projects = append(projects, ext)
			}
		}
	}
	if len(projects) == 0 {
		gp := detect.GenericPreset()
		projects = []detect.Project{{
			Name:   filepath.Base(p.Target),
			Path:   ".",
			Preset: &gp,
		}}
	}

	results := make([]counter.Result, 0, len(projects))
	var totalFiles, totalTokens int
	for _, proj := range projects {
		if !p.IncludeTests {
			proj = detect.WithoutTests(proj)
		}
		r, err := counter.Count(p.Target, proj, encoding)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %s: %v\n", proj.Name, err)
			continue
		}
		results = append(results, r)
		totalFiles += r.Files
		totalTokens += r.Tokens
	}

	return ScanResult{
		Results:       results,
		Projects:      projects,
		TotalFiles:    totalFiles,
		TotalTokens:   totalTokens,
		ContextWindow: ctxWin,
		Encoding:      encoding,
	}
}
