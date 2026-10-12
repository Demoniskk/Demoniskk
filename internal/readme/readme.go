// Package readme renders README.md from a text/template.
package readme

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"path"
	"path/filepath"
	"text/template"
	"time"

	"github.com/Demoniskk/Demoniskk/internal/card"
	"github.com/Demoniskk/Demoniskk/internal/config"
	"github.com/Demoniskk/Demoniskk/internal/stats"
)

type Data struct {
	User      *stats.Profile
	Config    *config.Config
	UpdatedAt time.Time
}

func Render(d Data) error {
	assets := filepath.ToSlash(d.Config.AssetsDir)
	// picture emits a <picture> that follows the viewer's GitHub theme.
	picture := func(name, alt, width string) string {
		dark := path.Join(assets, name+"-dark.svg")
		light := path.Join(assets, name+"-light.svg")
		return fmt.Sprintf(`<picture><source media="(prefers-color-scheme: dark)" srcset="%s"><source media="(prefers-color-scheme: light)" srcset="%s"><img alt="%s" src="%s" width="%s"></picture>`,
			dark, light, html.EscapeString(alt), dark, width)
	}
	funcs := template.FuncMap{
		"card": func(name, alt string) string { return picture(name, alt, "100%") },
		// repo emits a linked half-width tile; two sit side by side.
		"repo": func(r stats.Repo) string {
			return fmt.Sprintf(`<a href="%s">%s</a>`, r.URL, picture(card.RepoAsset(r.Name), r.Name+": "+r.Description, "49%"))
		},
		"num": card.FormatNum,
	}

	t, err := template.New(filepath.Base(d.Config.Template)).Funcs(funcs).ParseFiles(d.Config.Template)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, d); err != nil {
		return err
	}
	return os.WriteFile(d.Config.Output, buf.Bytes(), 0o644)
}
