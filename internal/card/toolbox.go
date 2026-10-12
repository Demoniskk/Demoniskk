package card

import (
	"strings"
	"unicode/utf8"

	"github.com/Demoniskk/Demoniskk/internal/config"
	"github.com/Demoniskk/Demoniskk/internal/stats"
)

// techColors gives familiar tools their brand color; anything else falls back to
// the language color GitHub reports, then to the accent.
var techColors = map[string]string{
	"go": "#00ADD8", "python": "#3572A5", "javascript": "#f1e05a", "typescript": "#3178c6",
	"html": "#e34c26", "css": "#663399", "html & css": "#e34c26", "rust": "#dea584",
	"c": "#a8b9cc", "c++": "#f34b7d", "c#": "#178600", "java": "#b07219", "lua": "#4c6ef5",
	"shell": "#89e051", "bash": "#89e051", "powershell": "#5391fe", "sql": "#e38c00",
	"git": "#f05032", "github": "#8b949e", "github actions": "#2088ff", "github pages": "#8b949e",
	"vs code": "#007acc", "docker": "#2496ed", "linux": "#fcc624", "windows": "#0078d4",
	"discord": "#5865f2", "steam": "#66c0f4", "node.js": "#5fa04e", "react": "#61dafb",
	"sqlite": "#0f80cc", "rss": "#f99b3a", "web crypto": "#a371f7", "claude code": "#d97757",
}

// Toolbox renders grouped tech pills.
func Toolbox(p *stats.Profile, groups []config.ToolGroup, s Style) string {
	const (
		x0, pillX = 32, 176
		pillH     = 28
		rowH      = 38
		groupGap  = 10
		charW     = 7.3
	)
	maxX := float64(width - x0)
	langColors := map[string]string{}
	for _, l := range p.Languages {
		langColors[strings.ToLower(l.Name)] = l.Color
	}

	// Lay out first so the card height is known before drawing.
	type pill struct {
		name, color string
		x           float64
		y           int
	}
	type label struct {
		text string
		y    int
	}
	var pills []pill
	var labels []label
	y := 80
	for gi, g := range groups {
		labels = append(labels, label{g.Group, y})
		x := float64(pillX)
		for _, item := range g.Items {
			w := float64(utf8.RuneCountInString(item))*charW + 34
			if x+w > maxX && x > pillX {
				x = pillX
				y += rowH
			}
			col := techColors[strings.ToLower(item)]
			if col == "" {
				col = langColors[strings.ToLower(item)]
			}
			if col == "" {
				col = mix(s.A1, s.A2, float64(len(pills)%4)/3)
			}
			pills = append(pills, pill{item, col, x, y})
			x += w + 8
		}
		if gi < len(groups)-1 {
			y += rowH + groupGap
		}
	}
	h := y + pillH + 32

	c := newCanvas(width, h, s, "Toolbox", "")
	c.f(`<text class="title fade" x="%d" y="50">Toolbox</text>`, x0)
	c.f(`<text class="mono muted sub fade" x="%d" y="50" text-anchor="end"%s>what I reach for</text>`, width-x0, delay(.1))
	for i, l := range labels {
		if i > 0 {
			c.f(`<rect x="%d" y="%d" width="%d" height="1" fill="%s" opacity=".6"/>`, x0, l.y-(rowH+groupGap)/2-6, width-2*x0, s.Border)
		}
		c.f(`<rect x="%d" y="%d" width="10" height="2" rx="1" fill="url(#accent)"/>`, x0, l.y+pillH/2-1)
		c.f(`<text class="label muted fade" x="%d" y="%d"%s>%s</text>`, x0+16, l.y+pillH/2+4, delay(float64(i)*.1), esc(strings.ToUpper(l.text)))
	}
	for i, pl := range pills {
		w := float64(utf8.RuneCountInString(pl.name))*charW + 34
		c.f(`<g class="fade"%s>`, delay(.15+float64(i)*.04))
		c.f(`<rect x="%.1f" y="%d" width="%.1f" height="%d" rx="%d" fill="%s" fill-opacity=".55" stroke="%s"/>`, pl.x, pl.y, w, pillH, pillH/2, s.Bg, s.Border)
		c.f(`<circle cx="%.1f" cy="%d" r="4.5" fill="%s"/>`, pl.x+15, pl.y+pillH/2, pl.color)
		c.f(`<text x="%.1f" y="%.1f" font-size="13" font-weight="500">%s</text>`, pl.x+26, float64(pl.y+pillH/2)+4.5, esc(pl.name))
		c.f(`</g>`)
	}
	return c.close()
}
