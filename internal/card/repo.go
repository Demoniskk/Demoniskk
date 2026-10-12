package card

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/Demoniskk/Demoniskk/internal/stats"
)

const tileW, tileH = 412, 164

var unsafeChars = regexp.MustCompile(`[^a-z0-9._-]+`)

// RepoAsset is the base file name for a repo tile, without the theme suffix.
func RepoAsset(name string) string {
	return "repo-" + unsafeChars.ReplaceAllString(strings.ToLower(name), "-")
}

// Repo renders a half-width tile for a featured repository.
func Repo(r stats.Repo, owner string, s Style) string {
	const x0 = 24
	c := newCanvas(tileW, tileH, s, r.Name, `.rname{font-size:17px;font-weight:700;letter-spacing:-.2px}
.arrow{fill:url(#accent)}`)

	c.f(`<text class="mono muted fade" x="%d" y="34" font-size="11.5">%s /</text>`, x0, esc(strings.ToLower(owner)))
	c.f(`<text class="rname fade" x="%d" y="56"%s>%s</text>`, x0, delay(.05), esc(truncate(r.Name, 32)))
	c.f(`<text class="arrow fade" x="%d" y="42" font-size="18" text-anchor="end"%s>↗</text>`, tileW-x0, delay(.1))

	desc := r.Description
	if desc == "" {
		desc = "No description yet."
	}
	for i, line := range wrap(desc, 54, 3) {
		c.f(`<text class="fade" x="%d" y="%d" font-size="13" opacity=".8"%s>%s</text>`, x0, 82+i*19, delay(.15+float64(i)*.05), esc(line))
	}

	// Footer: language, stars, live site.
	y := tileH - 20
	x := float64(x0)
	c.f(`<g class="fade"%s>`, delay(.35))
	if r.Language != "" {
		col := r.LanguageColor
		if col == "" {
			col = s.Muted
		}
		c.f(`<circle cx="%.1f" cy="%d" r="5" fill="%s"/>`, x+5, y-4, col)
		c.f(`<text x="%.1f" y="%d" font-size="12" class="muted">%s</text>`, x+16, y, esc(r.Language))
		x += 16 + float64(len(r.Language))*7 + 16
	}
	c.f(`<text x="%.1f" y="%d" font-size="12" class="muted">★ %s</text>`, x, y, FormatNum(r.Stars))
	x += 24 + float64(len(FormatNum(r.Stars)))*7
	if r.Forks > 0 {
		c.f(`<text x="%.1f" y="%d" font-size="12" class="muted">⑂ %s</text>`, x, y, FormatNum(r.Forks))
		x += 24 + float64(len(FormatNum(r.Forks)))*7
	}
	// The live site goes on the right, if it fits next to the stats.
	if host := siteHost(r.Homepage); host != "" {
		dotX := float64(tileW-x0) - float64(len([]rune(host)))*6.9 - 12
		if dotX-8 > x {
			c.f(`<circle class="pulse" cx="%.1f" cy="%d" r="3.5" fill="#3fb950"/>`, dotX, y-4)
			c.f(`<text class="mono muted" x="%d" y="%d" font-size="11.5" text-anchor="end">%s</text>`, tileW-x0, y, esc(host))
		}
	}
	c.f(`</g>`)
	return c.close()
}

// siteHost turns "https://demoniskk.github.io/nrk-rss/" into "demoniskk.github.io/nrk-rss".
func siteHost(raw string) string {
	u, err := url.Parse(raw)
	if raw == "" || err != nil || u.Host == "" {
		return ""
	}
	return truncate(strings.TrimSuffix(u.Host+u.Path, "/"), 30)
}
