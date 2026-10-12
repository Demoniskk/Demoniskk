// Package config loads the generator settings from config.json.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

type Link struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type Config struct {
	// Username is the GitHub login to render. Falls back to $GITHUB_REPOSITORY_OWNER.
	Username string `json:"username"`
	// Accent is the two-stop gradient used across every card.
	Accent []string `json:"accent"`
	// Status is shown as a pill in the corner of the profile card. Empty hides it.
	Status string `json:"status"`
	// Tagline replaces your GitHub bio on the profile card when set.
	Tagline string `json:"tagline"`

	TopLanguages     int      `json:"topLanguages"`
	ExcludeLanguages []string `json:"excludeLanguages"`
	ExcludeRepos     []string `json:"excludeRepos"`
	// IncludePrivate counts private repos in language stats (needs a PAT).
	IncludePrivate bool `json:"includePrivate"`

	AssetsDir string `json:"assetsDir"`
	Template  string `json:"template"`
	Output    string `json:"output"`

	About  []string `json:"about"`
	Skills []string `json:"skills"`
	Links  []Link   `json:"links"`
}

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func Load(path string) (*Config, error) {
	c := &Config{
		Accent:       []string{"#ff4d6d", "#a371f7"},
		TopLanguages: 8,
		AssetsDir:    "assets",
		Template:     "templates/README.md.tmpl",
		Output:       "README.md",
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(c.Accent) != 2 {
		return nil, fmt.Errorf("accent must list exactly 2 colors, got %d", len(c.Accent))
	}
	for _, col := range c.Accent {
		if !hexColor.MatchString(col) {
			return nil, fmt.Errorf("accent color %q must look like #rrggbb", col)
		}
	}
	c.TopLanguages = min(max(c.TopLanguages, 1), 10)
	return c, nil
}
