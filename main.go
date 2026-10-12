// Command Demoniskk generates the profile README and its animated SVG stat cards.
//
//	go run .            # fetch live stats (needs GH_STATS_TOKEN or GITHUB_TOKEN)
//	go run . -demo      # preview with sample data, no token needed
package main

import (
	"cmp"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Demoniskk/Demoniskk/internal/card"
	"github.com/Demoniskk/Demoniskk/internal/config"
	"github.com/Demoniskk/Demoniskk/internal/readme"
	"github.com/Demoniskk/Demoniskk/internal/stats"
)

func main() {
	cfgPath := flag.String("config", "config.json", "path to the config file")
	demo := flag.Bool("demo", false, "render with sample data instead of calling the GitHub API")
	user := flag.String("user", "", "GitHub login (overrides config and $GITHUB_REPOSITORY_OWNER)")
	flag.Parse()
	log.SetFlags(0)

	if err := run(*cfgPath, *user, *demo); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run(cfgPath, userFlag string, demo bool) error {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	login := cmp.Or(userFlag, cfg.Username, os.Getenv("GITHUB_REPOSITORY_OWNER"))
	now := time.Now().UTC()

	var p *stats.Profile
	if demo {
		p = stats.Demo(cmp.Or(login, "octocat"), now)
	} else {
		if login == "" {
			return fmt.Errorf("no username: set it in %s or pass -user", cfgPath)
		}
		token := cmp.Or(os.Getenv("GH_STATS_TOKEN"), os.Getenv("GITHUB_TOKEN"), os.Getenv("GH_TOKEN"))
		if token == "" {
			return fmt.Errorf("no token: set GH_STATS_TOKEN or GITHUB_TOKEN (or use -demo)")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		p, err = stats.Fetch(ctx, stats.NewClient(token), stats.Options{
			Login:            login,
			IncludePrivate:   cfg.IncludePrivate,
			TopLanguages:     cfg.TopLanguages,
			ExcludeLanguages: cfg.ExcludeLanguages,
			ExcludeRepos:     cfg.ExcludeRepos,
			Featured:         featuredNames(cfg),
		})
		if err != nil {
			return err
		}
	}
	for i, r := range p.Featured {
		for _, f := range cfg.Featured {
			if strings.EqualFold(f.Repo, r.Name) && f.Description != "" {
				p.Featured[i].Description = f.Description
			}
		}
	}

	if err := os.MkdirAll(cfg.AssetsDir, 0o755); err != nil {
		return err
	}
	// Drop tiles for repos that are no longer featured.
	stale, _ := filepath.Glob(filepath.Join(cfg.AssetsDir, "repo-*.svg"))
	for _, f := range stale {
		os.Remove(f)
	}
	for _, t := range []card.Theme{card.Dark, card.Light} {
		s := card.Style{Theme: t, Accent: cfg.Accent}
		cards := []struct{ name, svg string }{
			{"profile", card.Profile(p, s, cfg.Status, cfg.Tagline)},
			{"about", card.About(p.Login, cfg.AboutCommand, cfg.About, s)},
			{"toolbox", card.Toolbox(p, cfg.Toolbox, s)},
			{"languages", card.Languages(p, s)},
			{"activity", card.Activity(p, s)},
		}
		for _, r := range p.Featured {
			cards = append(cards, struct{ name, svg string }{card.RepoAsset(r.Name), card.Repo(r, p.Login, s)})
		}
		for _, c := range cards {
			out := filepath.Join(cfg.AssetsDir, c.name+"-"+t.Name+".svg")
			if err := os.WriteFile(out, []byte(c.svg), 0o644); err != nil {
				return err
			}
		}
	}

	if err := readme.Render(readme.Data{User: p, Config: cfg, UpdatedAt: now}); err != nil {
		return fmt.Errorf("render readme: %w", err)
	}

	log.Printf("✓ %s — %d commits, %d stars, %d languages, %d-day streak → %s + %s/",
		p.Login, p.Commits, p.Stars, len(p.Languages), p.CurrentStreak, cfg.Output, cfg.AssetsDir)
	return nil
}

func featuredNames(cfg *config.Config) []string {
	names := make([]string, len(cfg.Featured))
	for i, f := range cfg.Featured {
		names[i] = f.Repo
	}
	return names
}
