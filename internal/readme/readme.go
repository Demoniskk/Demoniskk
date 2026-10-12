// Package readme renders README.md from a text/template.
package readme

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
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
	funcs := template.FuncMap{
		// card emits a <picture> that follows the viewer's GitHub theme.
		"card": func(name, alt string) string {
			dark := path.Join(assets, name+"-dark.svg")
			light := path.Join(assets, name+"-light.svg")
			return fmt.Sprintf(`<picture>
  <source media="(prefers-color-scheme: dark)" srcset="%s">
  <source media="(prefers-color-scheme: light)" srcset="%s">
  <img alt="%s" src="%s" width="100%%">
</picture>`, dark, light, alt, dark)
		},
		"num": card.FormatNum,
		"short": func(s string, n int) string {
			if r := []rune(s); len(r) > n {
				return strings.TrimSpace(string(r[:n-1])) + "…"
			}
			return s
		},
		"skillicons": func(ids []string) string {
			return "https://skillicons.dev/icons?i=" + strings.Join(ids, ",") + "&perline=12"
		},
		// cell makes text safe inside a markdown table cell.
		"cell": func(s string) string {
			return strings.NewReplacer("|", `\|`, "\n", " ", "\r", "").Replace(s)
		},
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
