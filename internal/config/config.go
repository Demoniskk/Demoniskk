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

// AboutLine is one `key = value` row in the terminal-style about card.
// Text inside double quotes in Value is highlighted as a string.
type AboutLine struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ToolGroup struct {
	Group string   `json:"group"`
	Items []string `json:"items"`
}

type FeaturedRepo struct {
	Repo string `json:"repo"`
	// Description overrides the repo's own description when set.
	Description string `json:"description"`
}

// Font embeds a woff2 typeface into the cards. An empty Family uses system fonts.
type Font struct {
	Family    string `json:"family"`
	Regular   string `json:"regular"`
	Bold      string `json:"bold"`
	Monospace bool   `json:"monospace"`
}

type Config struct {
	// Username is the GitHub login to render. Falls back to $GITHUB_REPOSITORY_OWNER.
	Username string `json:"username"`
	// Accent is the single highlight color used sparingly across the cards.
	Accent string `json:"accent"`
	// Status is shown as a pill in the corner of the profile card. Empty hides it.
	Status string `json:"status"`
	// Tagline replaces your GitHub bio on the profile card when set.
	Tagline string `json:"tagline"`
	Font    Font   `json:"font"`
	// Timezone for the README's "last updated" line, as an IANA name.
	Timezone string `json:"timezone"`

	TopLanguages     int      `json:"topLanguages"`
	ExcludeLanguages []string `json:"excludeLanguages"`
	ExcludeRepos     []string `json:"excludeRepos"`
	// IncludePrivate counts private repos in language stats (needs a PAT).
	IncludePrivate bool `json:"includePrivate"`

	AssetsDir string `json:"assetsDir"`
	Template  string `json:"template"`
	Output    string `json:"output"`

	AboutCommand string         `json:"aboutCommand"`
	About        []AboutLine    `json:"about"`
	Toolbox      []ToolGroup    `json:"toolbox"`
	Featured     []FeaturedRepo `json:"featured"`
	Links        []Link         `json:"links"`
}

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func Load(path string) (*Config, error) {
	c := &Config{
		Accent:       "#e5534b",
		TopLanguages: 8,
		AssetsDir:    "assets",
		Template:     "templates/README.md.tmpl",
		Output:       "README.md",
		AboutCommand: "cat about.toml",
		Timezone:     "Europe/Oslo",
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if !hexColor.MatchString(c.Accent) {
		return nil, fmt.Errorf("accent color %q must look like #rrggbb", c.Accent)
	}
	c.TopLanguages = min(max(c.TopLanguages, 1), 10)
	return c, nil
}
