package stats

import (
	"hash/fnv"
	"math/rand/v2"
	"time"
)

// Demo returns believable sample data so the cards can be previewed without a token.
func Demo(login string, now time.Time) *Profile {
	h := fnv.New64a()
	h.Write([]byte(login))
	rng := rand.New(rand.NewPCG(h.Sum64(), 42))

	p := &Profile{
		Login: login, Name: "Demo " + login, Location: "Earth",
		Bio:       "Building fast, small, well-tested tools in Go. Currently exploring compilers, terminals and anything that renders pixels.",
		CreatedAt: now.AddDate(-5, -3, 0),
		Followers: 128, Following: 64, Repos: 37, Stars: 412, Forks: 58,
		PullRequests: 214, Issues: 96, ContributedTo: 23,
		Commits: 4821, CommitsThisYear: 1270, Contributions: 6390, ContributionsThisYear: 1544,
		Languages: []Language{
			{Name: "Go", Color: "#00ADD8", Percent: 41.2},
			{Name: "TypeScript", Color: "#3178c6", Percent: 22.7},
			{Name: "Rust", Color: "#dea584", Percent: 11.4},
			{Name: "Python", Color: "#3572A5", Percent: 8.9},
			{Name: "C", Color: "#555555", Percent: 6.1},
			{Name: "Shell", Color: "#89e051", Percent: 4.3},
			{Name: "Lua", Color: "#000080", Percent: 2.8},
			{Name: "Other", Color: "#8b949e", Percent: 2.6},
		},
		LanguageRepos: 31,
		Featured: []Repo{
			{Name: "fastcache", Description: "Sharded in-memory cache with zero allocations on the hot path", URL: "https://github.com/" + login, Language: "Go", Stars: 180},
			{Name: "glint", Description: "A tiny terminal UI toolkit", URL: "https://github.com/" + login, Language: "Rust", Stars: 96},
			{Name: "dotfiles", Description: "My setup, documented", URL: "https://github.com/" + login, Language: "Shell", Stars: 31},
		},
	}

	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	start := today.AddDate(0, 0, -364)
	start = start.AddDate(0, 0, -int(start.Weekday()))
	for d := start; !d.After(today); d = d.AddDate(0, 0, 1) {
		c := 0
		if rng.Float64() < 0.72 {
			c = 1 + rng.IntN(6)
			if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
				c += rng.IntN(8)
			}
		}
		p.Calendar = append(p.Calendar, Day{Date: d, Count: c, Weekday: int(d.Weekday())})
		p.ContributionsLastYear += c
	}
	p.finalize()
	return p
}
