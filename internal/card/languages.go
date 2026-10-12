package card

import (
	"fmt"

	"github.com/Demoniskk/Demoniskk/internal/stats"
)

// Languages renders a GitHub-style stacked bar plus a two-column breakdown.
func Languages(p *stats.Profile, s Style) string {
	const x0, barW, colGap, rowH, firstRow, barY = 20, width - 40, 40, 44, 112, 66
	langs := p.Languages
	rows := max((len(langs)+1)/2, 1)
	h := firstRow + rows*rowH - 6
	c := newCanvas(width, h, s, "Most used languages", "")
	c.header("Languages", fmt.Sprintf("by code size across %d repos", p.LanguageRepos))

	if len(langs) == 0 {
		c.f(`<text class="muted" x="%d" y="%d" text-anchor="middle" font-size="14">No language data yet</text>`, width/2, h/2+20)
		return c.close()
	}

	// Stacked bar.
	c.f(`<clipPath id="bar"><rect x="%d" y="%d" width="%d" height="8" rx="4"/></clipPath>`, x0, barY, barW)
	c.f(`<rect x="%d" y="%d" width="%d" height="8" rx="4" fill="%s"/>`, x0, barY, barW, s.Track)
	c.f(`<g clip-path="url(#bar)"><g class="grow"%s>`, delay(.15))
	pos := float64(x0)
	for i, l := range langs {
		w := l.Percent / 100 * barW
		if i < len(langs)-1 && w > 3 {
			w -= 2 // hairline gap between segments
		}
		c.f(`<rect x="%.2f" y="%d" width="%.2f" height="8" fill="%s"/>`, pos, barY, max(w, 0.5), l.Color)
		pos += l.Percent / 100 * barW
	}
	c.f(`</g></g>`)

	// Breakdown, filled column by column.
	colW := (barW - colGap) / 2
	top := langs[0].Percent
	for i, l := range langs {
		col, row := i/rows, i%rows
		x := x0 + col*(colW+colGap)
		y := firstRow + row*rowH
		d := .3 + float64(i)*.07
		c.f(`<g class="fade"%s>`, delay(d))
		c.f(`<circle cx="%d" cy="%d" r="4.5" fill="%s"/>`, x+5, y-5, l.Color)
		c.f(`<text x="%d" y="%d" font-size="14" font-weight="600">%s</text>`, x+18, y, esc(l.Name))
		c.f(`<text class="mono muted" x="%d" y="%d" font-size="12.5" text-anchor="end">%s</text>`, x+colW, y, pct(l.Percent))
		c.f(`<rect x="%d" y="%d" width="%d" height="4" rx="2" fill="%s"/>`, x, y+10, colW, s.Track)
		c.f(`</g>`)
		fill := max(l.Percent/top*float64(colW), 6)
		c.f(`<rect class="grow" x="%d" y="%d" width="%.2f" height="4" rx="2" fill="%s"%s/>`, x, y+10, fill, l.Color, delay(d+.1))
	}
	return c.close()
}

func pct(v float64) string {
	if v < 0.1 {
		return "<0.1%"
	}
	return fmt.Sprintf("%.1f%%", v)
}
