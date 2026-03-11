package counter

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/pkoukk/tiktoken-go"
	"github.com/ehmo/repo-tokens/internal/detect"
)

// FileTokens holds the token count for a single file.
type FileTokens struct {
	Path   string
	Tokens int
}

// Result holds the token counting results for a project.
type Result struct {
	Project    detect.Project
	Files      int
	Tokens     int
	SkippedBin int
	TopFiles   []FileTokens
}

// Count counts tokens in all matching files of a detected project.
func Count(root string, project detect.Project, encoding string) (Result, error) {
	enc, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		return Result{}, err
	}

	projectDir := filepath.Join(root, project.Path)
	skip := detect.MergeSkipDirs(project.Preset.SkipDirs)
	exts := detect.ExtensionSet(project.Preset.Extensions)

	files := collectFiles(projectDir, skip, exts, project.SkipTestDirSuffix, project.TestFileSuffixes)
	if len(files) == 0 {
		return Result{Project: project}, nil
	}

	workers := min(8, len(files))
	ch := make(chan string, len(files))
	for _, f := range files {
		ch <- f
	}
	close(ch)

	var totalTokens, skippedBin atomic.Int64
	var mu sync.Mutex
	var fileResults []FileTokens
	var wg sync.WaitGroup

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var local []FileTokens
			for path := range ch {
				data, err := os.ReadFile(path)
				if err != nil {
					continue
				}
				if isBinary(data) {
					skippedBin.Add(1)
					continue
				}
				n := len(enc.Encode(string(data), nil, nil))
				totalTokens.Add(int64(n))
				rel, _ := filepath.Rel(projectDir, path)
				local = append(local, FileTokens{Path: rel, Tokens: n})
			}
			mu.Lock()
			fileResults = append(fileResults, local...)
			mu.Unlock()
		}()
	}
	wg.Wait()

	sort.Slice(fileResults, func(i, j int) bool {
		return fileResults[i].Tokens > fileResults[j].Tokens
	})

	return Result{
		Project:    project,
		Files:      len(fileResults),
		Tokens:     int(totalTokens.Load()),
		SkippedBin: int(skippedBin.Load()),
		TopFiles:   fileResults,
	}, nil
}

func collectFiles(dir string, skip, exts map[string]bool, skipTestSuffix bool, testFileSuffixes []string) []string {
	var files []string
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if skip[name] || (strings.HasPrefix(name, ".") && name != ".") {
				return filepath.SkipDir
			}
			if skipTestSuffix && detect.IsTestDir(name) {
				return filepath.SkipDir
			}
			return nil
		}
		if detect.GlobalSkipFiles[d.Name()] {
			return nil
		}
		name := d.Name()
		if isGenerated(name) || isLockFile(name) || hasTestSuffix(name, testFileSuffixes) {
			return nil
		}
		if exts[strings.ToLower(filepath.Ext(path))] {
			files = append(files, path)
		}
		return nil
	})
	return files
}

func isBinary(data []byte) bool {
	check := data
	if len(check) > 8192 {
		check = check[:8192]
	}
	return bytes.IndexByte(check, 0) != -1
}

func isGenerated(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".min.js") ||
		strings.HasSuffix(lower, ".min.css") ||
		strings.HasSuffix(lower, ".min.mjs") ||
		strings.HasSuffix(lower, "-min.js") ||
		strings.HasSuffix(lower, "-min.css") ||
		strings.HasSuffix(lower, ".bundle.js") ||
		strings.HasSuffix(lower, ".chunk.js") ||
		lower == "output.css"
}

var lockFiles = map[string]bool{
	"package-lock.json": true, "yarn.lock": true, "pnpm-lock.yaml": true,
	"bun.lock": true, "composer.lock": true, "gemfile.lock": true,
	"poetry.lock": true, "cargo.lock": true, "go.sum": true,
}

func isLockFile(name string) bool {
	return lockFiles[strings.ToLower(name)]
}

func hasTestSuffix(name string, suffixes []string) bool {
	lower := strings.ToLower(name)
	for _, s := range suffixes {
		if strings.HasSuffix(lower, s) {
			return true
		}
	}
	return false
}
