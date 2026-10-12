package card

import (
	"cmp"
	"strings"

	"github.com/Demoniskk/Demoniskk/internal/config"
	"github.com/Demoniskk/Demoniskk/internal/stats"
)

// techColors gives familiar tools their brand color; anything else falls back to
// the language color GitHub reports, then to the muted text color.
var techColors = map[string]string{
	"go": "#00ADD8", "python": "#3572A5", "javascript": "#f1e05a", "typescript": "#3178c6",
	"html": "#e34c26", "css": "#663399", "html & css": "#e34c26", "rust": "#dea584",
	"c": "#a8b9cc", "c++": "#f34b7d", "c#": "#178600", "java": "#b07219", "lua": "#4c6ef5",
	"shell": "#89e051", "bash": "#89e051", "powershell": "#5391fe", "sql": "#e38c00",
	"git": "#f05032", "github actions": "#2088ff", "vs code": "#007acc", "docker": "#2496ed",
	"linux": "#fcc624", "windows": "#0078d4", "discord": "#5865f2", "discord api": "#5865f2",
	"steam": "#66c0f4", "steam web api": "#66c0f4", "node.js": "#5fa04e", "react": "#61dafb",
	"sqlite": "#0f80cc", "rss": "#f99b3a", "leetify": "#f84982", "claude code": "#d97757",
}

// Toolbox renders one column per group, each a plain list.
func Toolbox(p *stats.Profile, groups []config.ToolGroup, s Style) string {
	const (
		x0       = 20
		headY    = 74
		firstRow = 108
		rowH     = 26
	)
	langColors := map[string]string{}
	for _, l := range p.Languages {
		langColors[strings.ToLower(l.Name)] = l.Color
	}
	rows := 0
	for _, g := range groups {
		rows = max(rows, len(g.Items))
	}
	h := firstRow + rows*rowH - 4
	c := newCanvas(width, h, s, "Toolbox", "")
	c.header("Toolbox", "")

	colW := (width - 2*x0) / max(len(groups), 1)
	for gi, g := range groups {
		x := x0 + gi*colW
		c.f(`<g class="fade"%s>`, delay(float64(gi)*.06))
		c.f(`<text class="label muted" x="%d" y="%d">%s</text>`, x, headY, esc(strings.ToUpper(g.Group)))
		c.f(`<rect x="%d" y="%d" width="%d" height="1" fill="%s"/>`, x, headY+10, colW-24, s.Border)
		for i, item := range g.Items {
			key := strings.ToLower(item)
			col := cmp.Or(techColors[key], langColors[key], s.Muted)
			y := firstRow + i*rowH
			c.f(`<rect x="%d" y="%d" width="8" height="8" rx="2" fill="%s"/>`, x, y-9, col)
			c.f(`<text x="%d" y="%d" font-size="14">%s</text>`, x+17, y, esc(truncate(item, 22)))
		}
		c.f(`</g>`)
	}
	return c.close()
}
