// Package card renders animated SVG cards for the profile README.
//
// Everything is self-contained (no external fonts or images) because GitHub
// proxies README images through camo, which blocks nested requests.
package card

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Theme struct {
	Name    string
	Bg      string
	Surface string
	Border  string
	Text    string
	Muted   string
	Track   string
}

// Colors follow GitHub's Primer palette so the cards sit naturally on the page.
var (
	Dark  = Theme{Name: "dark", Bg: "#0d1117", Surface: "#151b23", Border: "#3d444d", Text: "#e6edf3", Muted: "#8d96a0", Track: "#1f252d"}
	Light = Theme{Name: "light", Bg: "#ffffff", Surface: "#f6f8fa", Border: "#d1d9e0", Text: "#1f2328", Muted: "#59636e", Track: "#eff2f5"}
)

type Style struct {
	Theme
	Accent string
}

const (
	width    = 840
	sansFont = `-apple-system,BlinkMacSystemFont,'Segoe UI','Noto Sans',Helvetica,Arial,sans-serif`
	monoFont = `ui-monospace,SFMono-Regular,'SF Mono',Menlo,Consolas,'Liberation Mono',monospace`
)

type canvas struct {
	strings.Builder
	w, h int
	s    Style
}

func (c *canvas) f(format string, args ...any) { fmt.Fprintf(c, format, args...) }

// newCanvas writes the shared frame: a flat box with a hairline border, like
// GitHub's own containers.
func newCanvas(w, h int, s Style, title, css string) *canvas {
	c := &canvas{w: w, h: h, s: s}
	c.f(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" fill="none" role="img" aria-label="%s">`, w, h, w, h, esc(title))
	c.f(`<title>%s</title>`, esc(title))
	c.f(`<defs><clipPath id="box"><rect x="1" y="1" width="%d" height="%d" rx="5.5"/></clipPath></defs>`, w-2, h-2)
	c.f(`<style>%s%s</style>`, baseCSS(s), css)
	c.f(`<rect x=".5" y=".5" width="%d" height="%d" rx="6" fill="%s" stroke="%s"/>`, w-1, h-1, s.Bg, s.Border)
	return c
}

// header draws a GitHub-style box header: a tinted strip with a title and an
// optional note on the right. It returns the y where content can start.
func (c *canvas) header(title, note string) int {
	const hh = 44
	c.f(`<rect x="1" y="1" width="%d" height="%d" fill="%s" clip-path="url(#box)"/>`, c.w-2, hh, c.s.Surface)
	c.f(`<rect x="1" y="%d" width="%d" height="1" fill="%s"/>`, hh+1, c.w-2, c.s.Border)
	c.f(`<text x="20" y="28" font-size="14" font-weight="600">%s</text>`, esc(title))
	if note != "" {
		c.f(`<text class="mono muted" x="%d" y="28" font-size="12" text-anchor="end">%s</text>`, c.w-20, esc(note))
	}
	return hh + 1
}

func (c *canvas) close() string {
	c.f(`</svg>`)
	return c.String()
}

func baseCSS(s Style) string {
	return fmt.Sprintf(`text{font-family:%s;fill:%s}
.mono{font-family:%s}
.muted{fill:%s}
.accent{fill:%s}
.label{font-size:11px;font-weight:600;letter-spacing:.8px}
.fade{animation:fade .6s ease-out both}
@keyframes fade{from{opacity:0}}
.grow{transform-box:fill-box;transform-origin:left center;animation:grow 1s cubic-bezier(.2,.8,.2,1) both}
@keyframes grow{from{transform:scaleX(0)}}
@media (prefers-reduced-motion:reduce){*{animation:none!important}}`,
		sansFont, s.Text, monoFont, s.Muted, s.Accent)
}

// stat draws a big number with a small caps label beneath.
func (c *canvas) stat(x, y int, value, unit, label string, delaySec float64) {
	c.f(`<g class="fade"%s>`, delay(delaySec))
	c.f(`<text x="%d" y="%d" font-size="22" font-weight="700">%s`, x, y, esc(value))
	if unit != "" {
		c.f(`<tspan class="muted" font-size="12.5" font-weight="500" dx="5">%s</tspan>`, esc(unit))
	}
	c.f(`</text>`)
	c.f(`<text class="label muted" x="%d" y="%d">%s</text>`, x, y+19, esc(strings.ToUpper(label)))
	c.f(`</g>`)
}

func delay(sec float64) string {
	if sec <= 0 {
		return ""
	}
	return fmt.Sprintf(` style="animation-delay:%.2fs"`, sec)
}

func esc(s string) string { return html.EscapeString(s) }

// FormatNum renders 1234 as "1,234" and 12345 as "12.3k".
func FormatNum(n int) string {
	switch {
	case n >= 1_000_000:
		return trimZero(fmt.Sprintf("%.1f", float64(n)/1_000_000)) + "m"
	case n >= 10_000:
		return trimZero(fmt.Sprintf("%.1f", float64(n)/1_000)) + "k"
	}
	s := strconv.Itoa(n)
	if len(s) > 3 {
		s = s[:len(s)-3] + "," + s[len(s)-3:]
	}
	return s
}

func trimZero(s string) string { return strings.TrimSuffix(s, ".0") }

// wrap splits text into at most maxLines lines of roughly maxChars, adding an
// ellipsis when it has to cut. SVG text has no native wrapping.
func wrap(s string, maxChars, maxLines int) []string {
	var lines []string
	var cur string
	for _, word := range strings.Fields(s) {
		switch {
		case cur == "":
			cur = word
		case utf8.RuneCountInString(cur)+1+utf8.RuneCountInString(word) <= maxChars:
			cur += " " + word
		default:
			lines = append(lines, cur)
			cur = word
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		last := []rune(lines[maxLines-1])
		if len(last) > maxChars-1 {
			last = last[:maxChars-1]
		}
		lines[maxLines-1] = strings.TrimRight(string(last), " .,;") + "…"
	}
	return lines
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return strings.TrimSpace(string(r[:n-1])) + "…"
}
