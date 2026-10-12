package card

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Demoniskk/Demoniskk/internal/stats"
)

// Activity renders the last year of contributions as an accent-tinted heatmap with streak stats.
func Activity(p *stats.Profile, s Style) string {
	const (
		h            = 290
		cell, step   = 11, 14
		gridX, gridY = 62, 86
		x0           = 20
	)

	// Column index per day; a new column starts every Sunday.
	cols := make([]int, len(p.Calendar))
	col := 0
	for i, d := range p.Calendar {
		if i > 0 && d.Weekday == 0 {
			col++
		}
		cols[i] = col
	}
	nCols := col + 1

	var css strings.Builder
	css.WriteString(`.c{transform-box:fill-box;transform-origin:center;animation:pop .5s cubic-bezier(.2,.8,.2,1) both}
@keyframes pop{from{opacity:0;transform:scale(.2)}}`)
	for i := range nCols {
		fmt.Fprintf(&css, ".d%d{animation-delay:%.3fs}", i, .2+float64(i)*.022)
	}
	c := newCanvas(width, h, s, "Contribution activity", css.String())

	c.header("Contributions", FormatNum(p.ContributionsLastYear)+" in the last year")

	// Month and weekday labels.
	lastMonth, lastLabelCol := -1, -10
	for i, d := range p.Calendar {
		if i > 0 && d.Weekday != 0 {
			continue
		}
		if m := int(d.Date.Month()); m != lastMonth {
			if cols[i]-lastLabelCol >= 3 && cols[i] < nCols-1 {
				c.f(`<text class="mono muted" x="%d" y="%d" font-size="10.5">%s</text>`, gridX+cols[i]*step, gridY-10, d.Date.Format("Jan"))
				lastLabelCol = cols[i]
			}
			lastMonth = m
		}
	}
	for row, name := range []string{1: "Mon", 3: "Wed", 5: "Fri"} {
		if name != "" {
			c.f(`<text class="mono muted" x="%d" y="%d" font-size="10">%s</text>`, x0, gridY+row*step+9, name)
		}
	}

	// Cells.
	levels := levelFunc(p.Calendar)
	for i, d := range p.Calendar {
		x, y := gridX+cols[i]*step, gridY+d.Weekday*step
		lvl := levels(d.Count)
		fill, op := levelColor(s, lvl)
		c.f(`<rect class="c d%d" x="%d" y="%d" width="%d" height="%d" rx="2.5" fill="%s"%s><title>%s: %d</title></rect>`,
			cols[i], x, y, cell, cell, fill, op, d.Date.Format("Jan 2, 2006"), d.Count)
	}
	if n := len(p.Calendar); n > 0 {
		last := p.Calendar[n-1]
		c.f(`<rect x="%.1f" y="%.1f" width="13" height="13" rx="3.5" stroke="%s"/>`,
			float64(gridX+cols[n-1]*step)-1, float64(gridY+last.Weekday*step)-1, s.Text)
	}

	// Legend.
	lx := width - x0 - 5*step - 30
	c.f(`<g class="fade"%s>`, delay(1.2))
	c.f(`<text class="mono muted" x="%d" y="%d" font-size="10" text-anchor="end">less</text>`, lx-6, gridY+7*step+16)
	for l := range 5 {
		fill, op := levelColor(s, l)
		c.f(`<rect x="%d" y="%d" width="%d" height="%d" rx="2.5" fill="%s"%s/>`, lx+l*step, gridY+7*step+7, cell, cell, fill, op)
	}
	c.f(`<text class="mono muted" x="%d" y="%d" font-size="10">more</text>`, lx+5*step+3, gridY+7*step+16)
	c.f(`</g>`)

	// Stats.
	const footerY = 214
	c.f(`<rect x="1" y="%d" width="%d" height="%d" fill="%s" clip-path="url(#box)"/>`, footerY, width-2, h-footerY, s.Surface)
	c.f(`<rect x="1" y="%d" width="%d" height="1" fill="%s"/>`, footerY, width-2, s.Border)
	best := "—"
	if p.BestDay.Count > 0 {
		best = p.BestDay.Date.Format("Jan 2")
	}
	items := []struct{ value, unit, label string }{
		{FormatNum(p.CurrentStreak), plural(p.CurrentStreak, "day"), "Current streak"},
		{FormatNum(p.LongestStreak), plural(p.LongestStreak, "day"), "Longest streak"},
		{FormatNum(p.BestDay.Count), best, "Best day"},
		{FormatNum(p.ContributionsThisYear), "", fmt.Sprintf("In %d", lastYear(p))},
		{FormatNum(p.Contributions), "", "All time"},
	}
	colW := (width - 2*x0) / len(items)
	for i, it := range items {
		c.stat(x0+i*colW, footerY+34, it.value, it.unit, it.label, .5+float64(i)*.08)
	}
	return c.close()
}

// levelFunc buckets counts into 0-4 using quartiles of the non-zero days,
// like GitHub, so one huge day doesn't wash out the rest of the year.
func levelFunc(days []stats.Day) func(int) int {
	var nz []int
	for _, d := range days {
		if d.Count > 0 {
			nz = append(nz, d.Count)
		}
	}
	slices.Sort(nz)
	q := func(f float64) int {
		if len(nz) == 0 {
			return 0
		}
		return nz[int(f*float64(len(nz)-1))]
	}
	q1, q2, q3 := q(.25), q(.5), q(.75)
	return func(n int) int {
		switch {
		case n == 0:
			return 0
		case n <= q1:
			return 1
		case n <= q2:
			return 2
		case n <= q3:
			return 3
		}
		return 4
	}
}

func levelColor(s Style, lvl int) (fill, opacity string) {
	if lvl == 0 {
		return s.Track, ""
	}
	return s.Accent, fmt.Sprintf(` fill-opacity="%.2f"`, []float64{0, .3, .5, .75, 1}[lvl])
}

func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}

func lastYear(p *stats.Profile) int {
	if n := len(p.Calendar); n > 0 {
		return p.Calendar[n-1].Date.Year()
	}
	return 0
}
