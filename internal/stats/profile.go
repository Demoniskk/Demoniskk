// Package stats collects profile statistics from the GitHub GraphQL API.
package stats

import (
	"slices"
	"time"
)

type Day struct {
	Date    time.Time
	Count   int
	Weekday int // 0 = Sunday
}

type Language struct {
	Name    string
	Color   string
	Bytes   int64
	Percent float64
}

type Repo struct {
	Name          string
	Description   string
	URL           string
	Language      string
	LanguageColor string
	Stars         int
	Forks         int
}

type Profile struct {
	Login     string
	Name      string
	Bio       string
	Location  string
	Company   string
	Website   string
	Twitter   string
	AvatarURL string
	Avatar    string // data URI, embedded so it renders inside the SVG
	CreatedAt time.Time

	Followers     int
	Following     int
	Repos         int
	Stars         int
	Forks         int
	PullRequests  int
	Issues        int
	ContributedTo int

	Commits               int // all time
	CommitsThisYear       int
	Contributions         int // all time
	ContributionsThisYear int
	ContributionsLastYear int // rolling 12 months, matches the calendar

	Calendar      []Day
	CurrentStreak int
	LongestStreak int
	BestDay       Day

	Languages     []Language
	LanguageRepos int
	Pinned        []Repo
}

func (p *Profile) DisplayName() string {
	if p.Name != "" {
		return p.Name
	}
	return p.Login
}

// finalize derives streaks and the best day from the calendar.
func (p *Profile) finalize() {
	slices.SortFunc(p.Calendar, func(a, b Day) int { return a.Date.Compare(b.Date) })

	run := 0
	p.LongestStreak, p.CurrentStreak, p.BestDay = 0, 0, Day{}
	for _, d := range p.Calendar {
		if d.Count > 0 {
			run++
			p.LongestStreak = max(p.LongestStreak, run)
		} else {
			run = 0
		}
		if d.Count > p.BestDay.Count {
			p.BestDay = d
		}
	}

	i := len(p.Calendar) - 1
	// Today isn't over yet, so an empty today doesn't break the streak.
	if i >= 0 && p.Calendar[i].Count == 0 {
		i--
	}
	for ; i >= 0 && p.Calendar[i].Count > 0; i-- {
		p.CurrentStreak++
	}
}
