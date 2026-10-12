package card

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Demoniskk/Demoniskk/internal/config"
)

// About renders a terminal window that types a command and prints key = value lines.
func About(login, command string, lines []config.AboutLine, s Style) string {
	const (
		x0         = 20
		promptY    = 80
		lineH      = 26
		charW      = 8.4 // approximate advance of a 14px monospace glyph
		maxValChar = 72
	)
	h := promptY + (len(lines)+1)*lineH + 26
	css := `.type{transform-box:fill-box;transform-origin:left center}
@keyframes type{from{transform:scaleX(0)}}
.cursor{animation:blink 1.1s steps(1) infinite}
@keyframes blink{50%{opacity:0}}
.t{font-size:14px}`
	c := newCanvas(width, h, s, "About "+login, css)

	// Title bar.
	c.f(`<rect x="1" y="1" width="%d" height="38" fill="%s" clip-path="url(#box)"/>`, width-2, s.Surface)
	c.f(`<rect x="1" y="39" width="%d" height="1" fill="%s"/>`, width-2, s.Border)
	for i := range 3 {
		c.f(`<circle cx="%d" cy="20" r="5" fill="%s"/>`, x0+4+i*17, s.Border)
	}
	c.f(`<text class="mono muted" x="%d" y="24" font-size="12" text-anchor="middle">%s@github: ~</text>`, width/2, esc(strings.ToLower(login)))

	// Typed command.
	n := utf8.RuneCountInString(command)
	typeDur := float64(n) * .05
	c.f(`<clipPath id="cmd"><rect class="type" x="%d" y="%d" width="%.0f" height="22" style="animation:type %.2fs steps(%d) .3s both"/></clipPath>`,
		x0+18, promptY-16, float64(n+1)*charW+4, typeDur, n)
	c.f(`<text class="mono t accent" x="%d" y="%d" font-weight="700">$</text>`, x0, promptY)
	c.f(`<text class="mono t" x="%d" y="%d" clip-path="url(#cmd)">%s</text>`, x0+18, promptY, esc(command))

	// Output, printed once the command has been "run".
	keyW := 0
	for _, l := range lines {
		keyW = max(keyW, utf8.RuneCountInString(l.Key))
	}
	eqX := float64(x0) + float64(keyW+1)*charW
	start := .3 + typeDur + .2
	for i, l := range lines {
		y := promptY + (i+1)*lineH
		c.f(`<g class="fade"%s>`, delay(start+float64(i)*.06))
		c.f(`<text class="mono t accent" x="%d" y="%d">%s</text>`, x0, y, esc(l.Key))
		c.f(`<text class="mono t muted" x="%.1f" y="%d">=</text>`, eqX, y)
		c.f(`<text class="mono t" x="%.1f" y="%d">%s</text>`, eqX+2*charW, y, highlight(truncate(l.Value, maxValChar)))
		c.f(`</g>`)
	}

	// Idle prompt with a blinking cursor.
	y := promptY + (len(lines)+1)*lineH
	c.f(`<g class="fade"%s>`, delay(start+float64(len(lines))*.06+.1))
	c.f(`<text class="mono t accent" x="%d" y="%d" font-weight="700">$</text>`, x0, y)
	c.f(`<rect class="cursor" x="%d" y="%d" width="8" height="16" fill="%s"/>`, x0+18, y-13, s.Text)
	c.f(`</g>`)
	return c.close()
}

// highlight leaves "quoted strings" in the text color and mutes quotes and punctuation.
func highlight(v string) string {
	var b strings.Builder
	for _, r := range v {
		if strings.ContainsRune(`[]{},"`, r) {
			fmt.Fprintf(&b, `<tspan class="muted">%s</tspan>`, esc(string(r)))
		} else {
			b.WriteString(esc(string(r)))
		}
	}
	return b.String()
}
