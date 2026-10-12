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
	Name        string
	Bg          string
	Surface     string
	Border      string
	Text        string
	Muted       string
	Track       string
	GlowOpacity float64
}

// Colors follow GitHub's Primer palette so the cards sit naturally on the page.
var (
	Dark  = Theme{Name: "dark", Bg: "#0d1117", Surface: "#161b22", Border: "#30363d", Text: "#e6edf3", Muted: "#8d96a0", Track: "#21262d", GlowOpacity: 0.22}
	Light = Theme{Name: "light", Bg: "#ffffff", Surface: "#f6f8fa", Border: "#d0d7de", Text: "#1f2328", Muted: "#59636e", Track: "#ebeef1", GlowOpacity: 0.14}
)

type Style struct {
	Theme
	A1, A2 string // accent gradient
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

// newCanvas writes the shared frame: gradient surface, dot grid, drifting glows
// and a shimmering accent edge along the top.
func newCanvas(w, h int, s Style, title, css string) *canvas {
	c := &canvas{w: w, h: h, s: s}
	c.f(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" fill="none" role="img" aria-label="%s">`, w, h, w, h, esc(title))
	c.f(`<title>%s</title><defs>`, esc(title))
	c.f(`<linearGradient id="accent" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s"/></linearGradient>`, s.A1, s.A2)
	c.f(`<linearGradient id="surface" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s"/></linearGradient>`, s.Bg, s.Surface)
	c.f(`<radialGradient id="glow1"><stop offset="0" stop-color="%s" stop-opacity="%.2f"/><stop offset="1" stop-color="%s" stop-opacity="0"/></radialGradient>`, s.A1, s.GlowOpacity, s.A1)
	c.f(`<radialGradient id="glow2"><stop offset="0" stop-color="%s" stop-opacity="%.2f"/><stop offset="1" stop-color="%s" stop-opacity="0"/></radialGradient>`, s.A2, s.GlowOpacity, s.A2)
	c.f(`<linearGradient id="edge" x1="0" y1="0" x2="1" y2="0"><stop offset="0" stop-color="%s" stop-opacity="0"/><stop offset=".25" stop-color="%s"/><stop offset=".75" stop-color="%s"/><stop offset="1" stop-color="%s" stop-opacity="0"/></linearGradient>`, s.A1, s.A1, s.A2, s.A2)
	c.f(`<linearGradient id="spark" x1="0" y1="0" x2="1" y2="0"><stop offset="0" stop-color="#fff" stop-opacity="0"/><stop offset=".5" stop-color="#fff" stop-opacity=".95"/><stop offset="1" stop-color="#fff" stop-opacity="0"/></linearGradient>`)
	c.f(`<pattern id="dots" width="20" height="20" patternUnits="userSpaceOnUse"><circle cx="2" cy="2" r="1" fill="%s" fill-opacity=".22"/></pattern>`, s.Muted)
	c.f(`<clipPath id="frame"><rect width="%d" height="%d" rx="16"/></clipPath>`, w, h)
	c.f(`</defs><style>%s%s</style>`, baseCSS(s, w), css)

	c.f(`<g clip-path="url(#frame)">`)
	c.f(`<rect width="%d" height="%d" fill="url(#surface)"/>`, w, h)
	c.f(`<rect width="%d" height="%d" fill="url(#dots)"/>`, w, h)
	c.f(`<circle class="drift" cx="%d" cy="-30" r="%d" fill="url(#glow1)"/>`, w-w/7, max(w/3, 150))
	c.f(`<circle class="drift2" cx="40" cy="%d" r="%d" fill="url(#glow2)"/>`, h+40, max(w*2/7, 130))
	c.f(`<rect width="%d" height="2" fill="url(#edge)"/>`, w)
	c.f(`<rect class="spark" x="-180" width="180" height="2" fill="url(#spark)"/>`)
	c.f(`</g>`)
	c.f(`<rect x=".5" y=".5" width="%d" height="%d" rx="15.5" stroke="%s"/>`, w-1, h-1, s.Border)
	return c
}

func (c *canvas) close() string {
	c.f(`</svg>`)
	return c.String()
}

func baseCSS(s Style, w int) string {
	return fmt.Sprintf(`text{font-family:%s;fill:%s}
.mono{font-family:%s}
.muted{fill:%s}
.title{font-size:18px;font-weight:600}
.sub{font-size:12px}
.label{font-size:10.5px;font-weight:600;letter-spacing:1.3px}
.fade{animation:fade .8s cubic-bezier(.2,.8,.2,1) both}
@keyframes fade{from{opacity:0;transform:translateY(6px)}to{opacity:1;transform:none}}
.grow{transform-box:fill-box;transform-origin:left center;animation:grow 1.2s cubic-bezier(.2,.8,.2,1) both}
@keyframes grow{from{transform:scaleX(0)}}
.drift{animation:drift 16s ease-in-out infinite alternate}
.drift2{animation:drift 21s ease-in-out infinite alternate-reverse}
@keyframes drift{from{transform:translate(0,0)}to{transform:translate(-90px,30px)}}
.spark{animation:spark 7s ease-in-out 1s infinite}
@keyframes spark{0%%{transform:translateX(0)}55%%,100%%{transform:translateX(%dpx)}}
.pulse{animation:pulse 2.4s ease-in-out infinite}
@keyframes pulse{0%%,100%%{opacity:1}50%%{opacity:.3}}
@media (prefers-reduced-motion:reduce){*{animation:none!important}}`,
		sansFont, s.Text, monoFont, s.Muted, w+180)
}

// stat draws a big number with an accent tick and a small caps label beneath.
func (c *canvas) stat(x, y int, value, unit, label string, delaySec float64) {
	c.f(`<g class="fade"%s>`, delay(delaySec))
	c.f(`<text x="%d" y="%d" font-size="22" font-weight="700">%s`, x, y, esc(value))
	if unit != "" {
		c.f(`<tspan class="muted" font-size="12.5" font-weight="500" dx="5">%s</tspan>`, esc(unit))
	}
	c.f(`</text>`)
	c.f(`<rect x="%d" y="%d" width="10" height="2" rx="1" fill="url(#accent)"/>`, x, y+13)
	c.f(`<text class="label muted" x="%d" y="%d">%s</text>`, x+16, y+17, esc(strings.ToUpper(label)))
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

// mix linearly blends two #rrggbb colors.
func mix(a, b string, t float64) string {
	pa, pb := parseHex(a), parseHex(b)
	var out [3]int
	for i := range out {
		out[i] = int(float64(pa[i])*(1-t) + float64(pb[i])*t + 0.5)
	}
	return fmt.Sprintf("#%02x%02x%02x", out[0], out[1], out[2])
}

func parseHex(h string) [3]int {
	v, _ := strconv.ParseUint(strings.TrimPrefix(h, "#"), 16, 32)
	return [3]int{int(v >> 16 & 0xff), int(v >> 8 & 0xff), int(v & 0xff)}
}
