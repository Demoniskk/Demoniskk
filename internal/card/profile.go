package card

import (
	"cmp"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Demoniskk/Demoniskk/internal/stats"
)

// Profile renders the hero card: avatar, name, bio, top languages and a footer of headline stats.
func Profile(p *stats.Profile, s Style, status, tagline string) string {
	const cx, r, x = 100, 60, 196
	bio := cmp.Or(tagline, p.Bio, fmt.Sprintf("%s contributions in the last year across %s repositories.",
		FormatNum(p.ContributionsLastYear), FormatNum(max(p.Repos, p.ContributedTo))))
	bioLines := wrap(bio, 72, 2)

	// Vertical rhythm depends on how many bio lines there are.
	chipsY := 114 + (len(bioLines)-1)*22 + 18
	footerY := chipsY + 24 + 26
	h := footerY + 50
	cy := footerY / 2

	c := newCanvas(width, h, s, p.DisplayName()+" on GitHub", "")

	// Avatar.
	c.f(`<clipPath id="av"><circle cx="%d" cy="%d" r="%d"/></clipPath>`, cx, cy, r)
	if p.Avatar != "" {
		c.f(`<image href="%s" x="%d" y="%d" width="%d" height="%d" clip-path="url(#av)" preserveAspectRatio="xMidYMid slice"/>`, p.Avatar, cx-r, cy-r, 2*r, 2*r)
	} else {
		initial, _ := utf8.DecodeRuneInString(strings.ToUpper(p.DisplayName()))
		c.f(`<circle cx="%d" cy="%d" r="%d" fill="%s"/>`, cx, cy, r, s.Track)
		c.f(`<text class="muted" x="%d" y="%d" text-anchor="middle" font-size="48" font-weight="600">%s</text>`, cx, cy+17, esc(string(initial)))
	}
	c.f(`<circle cx="%d" cy="%d" r="%.1f" stroke="%s"/>`, cx, cy, float64(r)+.5, s.Border)

	// Identity.
	c.f(`<text class="fade" x="%d" y="62" font-size="26" font-weight="600">%s</text>`, x, esc(truncate(p.DisplayName(), 28)))
	var meta []string
	if !strings.EqualFold(p.DisplayName(), p.Login) {
		meta = append(meta, p.Login)
	}
	if p.Location != "" {
		meta = append(meta, p.Location)
	}
	if !p.CreatedAt.IsZero() {
		meta = append(meta, fmt.Sprintf("on GitHub since %d", p.CreatedAt.Year()))
	}
	c.f(`<text class="muted fade" x="%d" y="86" font-size="14"%s>%s</text>`, x, delay(.05), esc(truncate(strings.Join(meta, " · "), 80)))
	for i, line := range bioLines {
		c.f(`<text class="fade" x="%d" y="%d" font-size="14.5"%s>%s</text>`, x, 114+i*22, delay(.1), esc(line))
	}
	languageChips(c, p.Languages, x, chipsY)

	if status != "" {
		statusPill(c, status)
	}

	// Footer strip with headline stats, like a GitHub box footer.
	c.f(`<rect x="1" y="%d" width="%d" height="%d" fill="%s" clip-path="url(#box)"/>`, footerY, width-2, h-footerY, s.Surface)
	c.f(`<rect x="1" y="%d" width="%d" height="1" fill="%s"/>`, footerY, width-2, s.Border)
	items := []struct {
		value int
		label string
	}{
		{p.Commits, "commits"},
		{p.PullRequests, "pull requests"},
		{p.Stars, "stars"},
		{p.Repos, "repositories"},
		{p.Followers, "followers"},
	}
	colW := (width - 40) / len(items)
	for i, it := range items {
		c.f(`<text class="fade" x="%d" y="%d" font-size="14"%s><tspan font-weight="600">%s</tspan><tspan class="muted" dx="5">%s</tspan></text>`,
			20+i*colW, footerY+30, delay(.2+float64(i)*.05), FormatNum(it.value), it.label)
	}
	return c.close()
}

// languageChips shows the top few languages as outlined labels. The dot and name
// are pinned left and the percentage right, so any difference between the
// estimated and real font width lands in the gap between them, never at an edge.
func languageChips(c *canvas, langs []stats.Language, x, y int) {
	const pad, nameX, gap = 9.0, 23.0, 6.0
	cx := float64(x)
	n := 0
	for _, l := range langs {
		if l.Name == "Other" || n == 4 {
			continue
		}
		pc := fmt.Sprintf("%.0f%%", l.Percent)
		w := nameX + textWidth(l.Name, 12) + gap + textWidth(pc, 12) + pad
		if cx+w > width-24 {
			break
		}
		c.f(`<g class="fade"%s>`, delay(.15))
		c.f(`<rect x="%.1f" y="%.1f" width="%.1f" height="23" rx="11.5" stroke="%s"/>`, cx+.5, float64(y)+.5, w, c.s.Border)
		c.f(`<circle cx="%.1f" cy="%.1f" r="4" fill="%s"/>`, cx+pad+4, float64(y)+12, l.Color)
		c.f(`<text x="%.1f" y="%d" font-size="12">%s</text>`, cx+nameX, y+16, esc(l.Name))
		c.f(`<text class="muted" x="%.1f" y="%d" font-size="12" text-anchor="end">%s</text>`, cx+w-pad+.5, y+16, pc)
		c.f(`</g>`)
		cx += w + 6
		n++
	}
}

func statusPill(c *canvas, status string) {
	status = truncate(status, 36)
	w := textWidth(status, 12) + 34
	x, y := float64(width-20)-w, 20.0
	c.f(`<g class="fade">`)
	c.f(`<rect x="%.1f" y="%.1f" width="%.1f" height="24" rx="12" stroke="%s"/>`, x+.5, y+.5, w, c.s.Border)
	c.f(`<circle cx="%.1f" cy="%.1f" r="4" fill="#3fb950"/>`, x+14, y+12)
	c.f(`<text class="muted" x="%.1f" y="%.1f" font-size="12">%s</text>`, x+25, y+16, esc(status))
	c.f(`</g>`)
}
