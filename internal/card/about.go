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
		x0         = 32
		promptY    = 92
		lineH      = 27
		charW      = 8.4 // approximate advance of a 14px monospace glyph
		maxValChar = 70
	)
	h := promptY + (len(lines)+1)*lineH + 34
	css := `.type{transform-box:fill-box;transform-origin:left center}
@keyframes type{from{transform:scaleX(0)}}
.cursor{animation:blink 1.1s steps(1) infinite}
@keyframes blink{50%{opacity:0}}
.t{font-size:14px}`
	c := newCanvas(width, h, s, "About "+login, css)

	// Title bar.
	for i, col := range []string{"#ff5f57", "#febc2e", "#28c840"} {
		c.f(`<circle cx="%d" cy="28" r="6" fill="%s"/>`, x0+i*20, col)
	}
	c.f(`<text class="mono muted" x="%d" y="32" font-size="12" text-anchor="middle">%s@github: ~</text>`, width/2, esc(strings.ToLower(login)))
	c.f(`<rect y="52" width="%d" height="1" fill="%s"/>`, width, s.Border)

	// Typed command.
	typeDur := float64(utf8.RuneCountInString(command)) * .055
	c.f(`<clipPath id="cmd"><rect class="type" x="%d" y="%d" width="%.0f" height="22" style="animation:type %.2fs steps(%d) .4s both"/></clipPath>`,
		x0+20, promptY-16, float64(utf8.RuneCountInString(command)+1)*charW+4, typeDur, utf8.RuneCountInString(command))
	c.f(`<text class="mono t" x="%d" y="%d" style="fill:url(#accent)" font-weight="700">❯</text>`, x0, promptY)
	c.f(`<text class="mono t" x="%d" y="%d" clip-path="url(#cmd)">%s</text>`, x0+20, promptY, esc(command))

	// Output, printed line by line once the command has been "run".
	keyW := 0
	for _, l := range lines {
		keyW = max(keyW, utf8.RuneCountInString(l.Key))
	}
	eqX := float64(x0) + float64(keyW+1)*charW
	start := .4 + typeDur + .25
	for i, l := range lines {
		y := promptY + (i+1)*lineH
		c.f(`<g class="fade"%s>`, delay(start+float64(i)*.12))
		c.f(`<text class="mono t" x="%d" y="%d" style="fill:%s">%s</text>`, x0, y, s.A1, esc(l.Key))
		c.f(`<text class="mono t muted" x="%.1f" y="%d">=</text>`, eqX, y)
		c.f(`<text class="mono t" x="%.1f" y="%d">%s</text>`, eqX+2*charW, y, highlight(truncate(l.Value, maxValChar), s))
		c.f(`</g>`)
	}

	// Idle prompt with a blinking cursor.
	y := promptY + (len(lines)+1)*lineH
	c.f(`<g class="fade"%s>`, delay(start+float64(len(lines))*.12+.1))
	c.f(`<text class="mono t" x="%d" y="%d" style="fill:url(#accent)" font-weight="700">❯</text>`, x0, y)
	c.f(`<rect class="cursor" x="%d" y="%d" width="8" height="17" rx="1" fill="%s"/>`, x0+20, y-13, s.A1)
	c.f(`</g>`)
	return c.close()
}

// highlight colors "quoted strings" with the second accent and punctuation as muted.
func highlight(v string, s Style) string {
	var b strings.Builder
	inStr := false
	for _, part := range strings.SplitAfter(v, `"`) {
		closed := strings.HasSuffix(part, `"`)
		body := strings.TrimSuffix(part, `"`)
		switch {
		case inStr:
			fmt.Fprintf(&b, `<tspan style="fill:%s">%s"</tspan>`, s.A2, esc(body))
		default:
			b.WriteString(punct(body))
			if closed {
				fmt.Fprintf(&b, `<tspan style="fill:%s">"</tspan>`, s.A2)
			}
		}
		if closed {
			inStr = !inStr
		}
	}
	return b.String()
}

func punct(s string) string {
	var b strings.Builder
	for _, r := range s {
		if strings.ContainsRune("[]{},", r) {
			fmt.Fprintf(&b, `<tspan class="muted">%s</tspan>`, esc(string(r)))
		} else {
			b.WriteString(esc(string(r)))
		}
	}
	return b.String()
}
