// Package card renders animated SVG cards for the profile README.
//
// Everything is self-contained (no external fonts or images) because GitHub
// proxies README images through camo, which blocks nested requests.
package card

import (
	"encoding/base64"
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

// Font is a typeface embedded into every card. Embedding is the only way to use
// a custom font, because GitHub's image proxy blocks external font requests.
type Font struct {
	Family    string
	Regular   []byte // woff2
	Bold      []byte // woff2, optional
	Monospace bool
}

type Style struct {
	Theme
	Accent string
	Font   *Font // nil uses the viewer's system fonts
}

func (s Style) mono() bool { return s.Font != nil && s.Font.Monospace }

// textWidth measures text in the card font: exact for a monospace font,
// a slightly generous estimate otherwise.
func (s Style) textWidth(text string, size float64) float64 {
	if s.mono() {
		return monoWidth(text, size)
	}
	return textWidth(text, size)
}

// charsFit is how many characters fit in maxW at size. sansChars is the tuned
// count used with proportional system fonts.
func (s Style) charsFit(size, maxW float64, sansChars int) int {
	if s.mono() {
		return int(maxW / (.6 * size))
	}
	return sansChars
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
	sans, mono, faces := sansFont, monoFont, ""
	if f := s.Font; f != nil {
		faces = fontFace(f.Family, 400, f.Regular)
		if len(f.Bold) > 0 {
			faces += fontFace(f.Family, 700, f.Bold)
		}
		sans = "'" + f.Family + "'," + sansFont
		mono = "'" + f.Family + "'," + monoFont
	}
	return faces + fmt.Sprintf(`text{font-family:%s;fill:%s}
.mono{font-family:%s}
.muted{fill:%s}
.accent{fill:%s}
.label{font-size:11px;font-weight:600;letter-spacing:.8px}
.fade{animation:fade .6s ease-out both}
@keyframes fade{from{opacity:0}}
.grow{transform-box:fill-box;transform-origin:left center;animation:grow 1s cubic-bezier(.2,.8,.2,1) both}
@keyframes grow{from{transform:scaleX(0)}}
@media (prefers-reduced-motion:reduce){*{animation:none!important}}`,
		sans, s.Text, mono, s.Muted, s.Accent)
}

func fontFace(family string, weight int, woff2 []byte) string {
	return fmt.Sprintf(`@font-face{font-family:'%s';font-weight:%d;src:url(data:font/woff2;base64,%s) format('woff2')}`,
		family, weight, base64.StdEncoding.EncodeToString(woff2))
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

// textWidth estimates how wide s renders in the sans stack at size px. It errs
// wide, so callers reserve this much space and center the text inside it; the
// viewer's real font then never overflows, it just gets a little extra padding.
func textWidth(s string, size float64) float64 {
	em := 0.0
	for _, r := range s {
		switch {
		case strings.ContainsRune("il.,:;'|!", r):
			em += .3
		case strings.ContainsRune("fjrt ", r):
			em += .38
		case r == 'm' || r == 'w':
			em += .86
		case r == 'M' || r == 'W':
			em += .9
		case r >= 'A' && r <= 'Z':
			em += .7
		case r >= '0' && r <= '9':
			em += .6
		case r == '%':
			em += .92
		default:
			em += .58
		}
	}
	return em * size * .94
}

// monoWidth is the width of s in a monospace font at size px (0.6em per glyph,
// exact for JetBrains Mono and close for common system monospace fonts).
func monoWidth(s string, size float64) float64 {
	return float64(utf8.RuneCountInString(s)) * .6 * size
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return strings.TrimSpace(string(r[:n-1])) + "…"
}
