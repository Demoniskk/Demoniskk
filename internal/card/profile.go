package card

import (
	"cmp"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Demoniskk/Demoniskk/internal/stats"
)

// Profile renders the hero card: avatar with orbiting rings, name, bio and headline stats.
func Profile(p *stats.Profile, s Style, status, tagline string) string {
	const h = 268
	const cx, cy, r = 130, 134, 74
	css := fmt.Sprintf(`.orbit{transform-origin:%[1]dpx %[2]dpx;animation:spin 18s linear infinite}
.orbit2{transform-origin:%[1]dpx %[2]dpx;animation:spin 40s linear infinite reverse}
@keyframes spin{to{transform:rotate(360deg)}}
.name{font-size:34px;font-weight:800;letter-spacing:-.6px}`, cx, cy)
	c := newCanvas(h, s, p.DisplayName()+" on GitHub", css)

	// Avatar.
	c.f(`<circle class="pulse" cx="%d" cy="%d" r="108" fill="url(#glow1)"/>`, cx, cy)
	c.f(`<circle class="orbit2" cx="%d" cy="%d" r="94" stroke="%s" stroke-opacity=".6" stroke-dasharray="1.5 7" stroke-linecap="round"/>`, cx, cy, s.Muted)
	c.f(`<circle class="orbit" cx="%d" cy="%d" r="84" stroke="url(#accent)" stroke-width="3" stroke-linecap="round" stroke-dasharray="118 46 30 46"/>`, cx, cy)
	c.f(`<clipPath id="av"><circle cx="%d" cy="%d" r="%d"/></clipPath>`, cx, cy, r)
	if p.Avatar != "" {
		c.f(`<image href="%s" x="%d" y="%d" width="%d" height="%d" clip-path="url(#av)" preserveAspectRatio="xMidYMid slice"/>`, p.Avatar, cx-r, cy-r, 2*r, 2*r)
	} else {
		initial, _ := utf8.DecodeRuneInString(strings.ToUpper(p.DisplayName()))
		c.f(`<circle cx="%d" cy="%d" r="%d" fill="url(#accent)"/>`, cx, cy, r)
		c.f(`<text x="%d" y="%d" text-anchor="middle" font-size="60" font-weight="800" style="fill:#fff">%s</text>`, cx, cy+21, esc(string(initial)))
	}
	c.f(`<circle cx="%d" cy="%d" r="%d" stroke="%s" stroke-width="2"/>`, cx, cy, r, s.Border)

	// Identity.
	x := 252
	c.f(`<text class="name fade" x="%d" y="88" style="fill:url(#accent)">%s</text>`, x, esc(truncate(p.DisplayName(), 24)))

	meta := []string{"@" + p.Login}
	if p.Location != "" {
		meta = append(meta, p.Location)
	}
	if !p.CreatedAt.IsZero() {
		meta = append(meta, fmt.Sprintf("since %d", p.CreatedAt.Year()))
	}
	c.f(`<text class="mono muted fade" x="%d" y="115" font-size="13"%s>%s</text>`, x, delay(.1), esc(truncate(strings.Join(meta, "  ·  "), 72)))

	bio := cmp.Or(tagline, p.Bio, fmt.Sprintf("%s contributions in the last year across %s repositories.",
		FormatNum(p.ContributionsLastYear), FormatNum(max(p.Repos, p.ContributedTo))))
	for i, line := range wrap(bio, 66, 2) {
		c.f(`<text class="fade" x="%d" y="%d" font-size="15" opacity=".92"%s>%s</text>`, x, 146+i*22, delay(.2+float64(i)*.05), esc(line))
	}

	// Headline stats.
	c.f(`<rect class="grow" x="%d" y="182" width="%d" height="1" fill="%s"%s/>`, x, width-x-32, s.Border, delay(.3))
	items := []struct {
		label string
		value int
	}{
		{"Commits", p.Commits},
		{"Stars", p.Stars},
		{"Pull reqs", p.PullRequests},
		{"Followers", p.Followers},
		{"Repos", p.Repos},
	}
	colW := (width - x - 32) / len(items)
	for i, it := range items {
		c.stat(x+i*colW, 224, FormatNum(it.value), "", it.label, .35+float64(i)*.08)
	}

	if status != "" {
		statusPill(c, status)
	}
	return c.close()
}

func statusPill(c *canvas, status string) {
	status = truncate(status, 36)
	w := int(float64(utf8.RuneCountInString(status))*7.25) + 40
	x, y := width-28-w, 24
	c.f(`<g class="fade"%s>`, delay(.5))
	c.f(`<rect x="%d" y="%d" width="%d" height="26" rx="13" fill="%s" fill-opacity=".7" stroke="%s"/>`, x, y, w, c.s.Bg, c.s.Border)
	c.f(`<circle class="pulse" cx="%d" cy="%d" r="7" fill="#3fb950" fill-opacity=".25"/>`, x+16, y+13)
	c.f(`<circle cx="%d" cy="%d" r="3.5" fill="#3fb950"/>`, x+16, y+13)
	c.f(`<text class="mono" x="%d" y="%d" font-size="12">%s</text>`, x+28, y+17, esc(status))
	c.f(`</g>`)
}
