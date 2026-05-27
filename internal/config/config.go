package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

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
	applyProgressLogFromEnv(&cfg)
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

func applyProgressLogFromEnv(cfg *Config) {
	repoURL := os.Getenv("PROGRESS_LOG_REPO_URL")
	if repoURL == "" {
		repoURL = os.Getenv("PROGRESS_LOG_REPO")
	}
	if repoURL != "" {
		owner, repo, err := parseGitHubRepoURL(repoURL)
		if err != nil {
			return
		}
		cfg.ProgressLog.Owner = owner
		cfg.ProgressLog.Repo = repo
	}

	if branch := os.Getenv("PROGRESS_LOG_BRANCH"); branch != "" {
		cfg.ProgressLog.Branch = branch
	}
	if path := os.Getenv("PROGRESS_LOG_PATH"); path != "" {
		cfg.ProgressLog.Path = path
	}
}

func parseGitHubRepoURL(raw string) (owner, repo string, err error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimSuffix(raw, ".git")
	raw = strings.TrimSuffix(raw, "/")

	const prefix = "github.com/"
	idx := strings.Index(raw, prefix)
	if idx < 0 {
		return "", "", fmt.Errorf("invalid github repo url: %s", raw)
	}

	path := strings.Trim(raw[idx+len(prefix):], "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid github repo url: %s", raw)
	}

	return parts[0], parts[1], nil
}

func (c *Config) applyProgressLogDefaults() {
	if c.ProgressLog.Owner == "" {
		c.ProgressLog.Owner = "aryan735"
	}
	if c.ProgressLog.Repo == "" {
		c.ProgressLog.Repo = "engineering-progress-log"
	}
	if c.ProgressLog.Branch == "" {
		c.ProgressLog.Branch = "main"
	}
	if c.ProgressLog.Path == "" {
		c.ProgressLog.Path = "engineering-progress-log.md"
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

func (c *Config) ProgressLogRepoURL() string {
	return fmt.Sprintf("https://github.com/%s/%s", c.ProgressLog.Owner, c.ProgressLog.Repo)
}
