package config

import (
	"os"
	"path/filepath"

	"github.com/ehmo/repo-tokens/internal/detect"
	"gopkg.in/yaml.v3"
)

// Config is the .repo-tokens.yml file format.
type Config struct {
	ContextWindow int       `yaml:"context-window"`
	Encoding      string    `yaml:"encoding"`
	Exclude       []string  `yaml:"exclude"`
	Projects      []Project `yaml:"projects"`
}

// Project defines a project override in the config file.
type Project struct {
	Name    string   `yaml:"name"`
	Path    string   `yaml:"path"`
	Type    string   `yaml:"type"`
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

// Load reads .repo-tokens.yml from dir. Returns nil if the file doesn't exist.
func Load(dir string) (*Config, error) {
	data, err := os.ReadFile(filepath.Join(dir, ".repo-tokens.yml"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.ContextWindow == 0 {
		cfg.ContextWindow = 200_000
	}
	if cfg.Encoding == "" {
		cfg.Encoding = "cl100k_base"
	}
	return &cfg, nil
}

// ToProjects converts config projects to detect.Projects.
func ToProjects(cfg *Config) []detect.Project {
	projects := make([]detect.Project, 0, len(cfg.Projects))
	for _, cp := range cfg.Projects {
		preset := detect.FindPreset(cp.Type)
		if preset == nil {
			preset = &detect.Preset{
				Name:       cp.Type,
				Extensions: cp.Include,
				SkipDirs:   cp.Exclude,
			}
		} else if len(cp.Exclude) > 0 {
			merged := *preset
			skipDirs := make([]string, 0, len(preset.SkipDirs)+len(cp.Exclude))
			skipDirs = append(skipDirs, preset.SkipDirs...)
			skipDirs = append(skipDirs, cp.Exclude...)
			merged.SkipDirs = skipDirs
			preset = &merged
		}
		projects = append(projects, detect.Project{Name: cp.Name, Path: cp.Path, Preset: preset})
	}
	return projects
}
