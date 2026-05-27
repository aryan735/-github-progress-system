package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/aryan735/-github-progress-system/internal/github"
	"gopkg.in/yaml.v3"
)

type ProgressLogConfig struct {
	Owner  string `yaml:"owner"`
	Repo   string `yaml:"repo"`
	Branch string `yaml:"branch"`
	Path   string `yaml:"path"`
}

type Config struct {
	GithubToken string            `yaml:"github_token"`
	ProgressLog ProgressLogConfig `yaml:"progress_log"`
}

func Load() (*Config, error) {
	return LoadConfig(os.Getenv("CONFIG_PATH"))
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config

	paths := []string{}
	if path != "" {
		paths = append(paths, path)
	}
	paths = append(paths,
		"internal/config/config.yml",
		"../../internal/config/config.yml",
	)

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse config %s: %w", p, err)
		}
		break
	}

	applyTokenFromEnv(&cfg)
	cfg.applyProgressLogDefaults()

	if cfg.GithubToken == "" {
		return nil, errors.New("github_token is not set (config file or GH_PROGRESS_TOKEN / github_token env)")
	}

	return &cfg, nil
}

func applyTokenFromEnv(cfg *Config) {
	if token := os.Getenv("github_token"); token != "" {
		cfg.GithubToken = token
		return
	}
	if token := os.Getenv("GH_PROGRESS_TOKEN"); token != "" {
		cfg.GithubToken = token
	}
}

func (c *Config) applyProgressLogDefaults() {
	if c.ProgressLog.Owner == "" {
		c.ProgressLog.Owner = "aryan735"
	}
	if c.ProgressLog.Repo == "" {
		c.ProgressLog.Repo = "-github-progress-system"
	}
	if c.ProgressLog.Branch == "" {
		c.ProgressLog.Branch = "main"
	}
	if c.ProgressLog.Path == "" {
		c.ProgressLog.Path = "docs/developer-progress-log.md"
	}
}

func (c *Config) ProgressLogTarget() github.ProgressLogTarget {
	return github.ProgressLogTarget{
		Owner:  c.ProgressLog.Owner,
		Repo:   c.ProgressLog.Repo,
		Branch: c.ProgressLog.Branch,
		Path:   c.ProgressLog.Path,
	}
}
