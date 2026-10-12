package card

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/Demoniskk/Demoniskk/internal/stats"
)

const tileW, tileH = 412, 150

var unsafeChars = regexp.MustCompile(`[^a-z0-9._-]+`)

// RepoAsset is the base file name for a repo tile, without the theme suffix.
func RepoAsset(name string) string {
	return "repo-" + unsafeChars.ReplaceAllString(strings.ToLower(name), "-")
}

// Repo renders a half-width tile for a featured repository.
func Repo(r stats.Repo, owner string, s Style) string {
	const x0 = 20
	c := newCanvas(tileW, tileH, s, r.Name, "")

	c.f(`<rect x="%d" y="22" width="3" height="16" rx="1" fill="%s"/>`, x0, s.Accent)
	c.f(`<text class="fade" x="%d" y="36" font-size="15" font-weight="600">%s</text>`, x0+11, esc(truncate(r.Name, 34)))
	c.f(`<text class="muted fade" x="%d" y="36" font-size="15" text-anchor="end">↗</text>`, tileW-x0)

	desc := r.Description
	if desc == "" {
		desc = "No description yet."
	}
	for i, line := range wrap(desc, 56, 3) {
		c.f(`<text class="muted fade" x="%d" y="%d" font-size="13"%s>%s</text>`, x0, 64+i*19, delay(.05), esc(line))
	}

	// Footer: language, stars, live site.
	y := tileH - 18
	x := float64(x0)
	c.f(`<g class="fade"%s>`, delay(.1))
	if r.Language != "" {
		col := r.LanguageColor
		if col == "" {
			col = s.Muted
		}
		c.f(`<circle cx="%.1f" cy="%d" r="4.5" fill="%s"/>`, x+5, y-4, col)
		lw := textWidth(r.Language, 12)
		c.f(`<text x="%.1f" y="%d" font-size="12" class="muted"%s>%s</text>`, x+16, y, fitText(lw), esc(r.Language))
		x += 16 + lw + 16
	}
	for _, st := range []struct {
		icon string
		n    int
	}{{"★", r.Stars}, {"⑂", r.Forks}} {
		if st.n == 0 && st.icon == "⑂" {
			continue
		}
		v := FormatNum(st.n)
		c.f(`<text x="%.1f" y="%d" font-size="12" class="muted">%s</text>`, x, y, st.icon)
		c.f(`<text x="%.1f" y="%d" font-size="12" class="muted"%s>%s</text>`, x+14, y, fitText(textWidth(v, 12)), v)
		x += 14 + textWidth(v, 12) + 16
	}
	// The live site goes on the right, if it fits next to the stats.
	if host := siteHost(r.Homepage); host != "" {
		hw := monoWidth(host, 11.5)
		dotX := float64(tileW-x0) - hw - 10
		if dotX-8 > x {
			c.f(`<circle cx="%.1f" cy="%d" r="3.5" fill="#3fb950"/>`, dotX, y-4)
			c.f(`<text class="mono muted" x="%d" y="%d" font-size="11.5" text-anchor="end"%s>%s</text>`, tileW-x0, y, fitText(hw), esc(host))
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
