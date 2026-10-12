package stats

import (
	"bytes"
	"cmp"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"
)

const endpoint = "https://api.github.com/graphql"

type Client struct {
	token string
	http  *http.Client
}

func NewClient(token string) *Client {
	return &Client{token: token, http: &http.Client{Timeout: 30 * time.Second}}
}

type Options struct {
	Login            string
	IncludePrivate   bool
	TopLanguages     int
	ExcludeLanguages []string
	ExcludeRepos     []string
}

const repoFields = `
fragment repo on Repository {
  name description url stargazerCount forkCount isPrivate isArchived
  primaryLanguage { name color }
}`

const userQuery = `query($login: String!) {
  user(login: $login) {
    login name bio location company websiteUrl twitterUsername createdAt
    avatarUrl(size: 256)
    followers { totalCount }
    following { totalCount }
    pullRequests { totalCount }
    issues { totalCount }
    repositoriesContributedTo(contributionTypes: [COMMIT, PULL_REQUEST, ISSUE, REPOSITORY]) { totalCount }
    pinnedItems(first: 6, types: REPOSITORY) { nodes { ...repo } }
    contributionsCollection {
      contributionYears
      contributionCalendar {
        totalContributions
        weeks { contributionDays { date contributionCount weekday } }
      }
    }
  }
}` + repoFields

const reposQuery = `query($login: String!, $cursor: String, $privacy: RepositoryPrivacy) {
  user(login: $login) {
    repositories(first: 100, after: $cursor, ownerAffiliations: OWNER, isFork: false, privacy: $privacy) {
      pageInfo { hasNextPage endCursor }
      nodes {
        ...repo
        languages(first: 20, orderBy: {field: SIZE, direction: DESC}) {
          edges { size node { name color } }
        }
      }
    }
  }
}` + repoFields

type count struct {
	TotalCount int `json:"totalCount"`
}

type repoNode struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	URL             string `json:"url"`
	StargazerCount  int    `json:"stargazerCount"`
	ForkCount       int    `json:"forkCount"`
	IsPrivate       bool   `json:"isPrivate"`
	IsArchived      bool   `json:"isArchived"`
	PrimaryLanguage *struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	} `json:"primaryLanguage"`
	Languages struct {
		Edges []struct {
			Size int64 `json:"size"`
			Node struct {
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"languages"`
}

func (r repoNode) toRepo() Repo {
	out := Repo{Name: r.Name, Description: r.Description, URL: r.URL, Stars: r.StargazerCount, Forks: r.ForkCount}
	if r.PrimaryLanguage != nil {
		out.Language, out.LanguageColor = r.PrimaryLanguage.Name, r.PrimaryLanguage.Color
	}
	return out
}

func (c *Client) query(ctx context.Context, q string, vars map[string]any, out any) error {
	body, err := json.Marshal(map[string]any{"query": q, "variables": vars})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "profile-readme-generator")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("github api: %s: %s", resp.Status, strings.TrimSpace(string(msg)))
	}

	var env struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if len(env.Errors) > 0 {
		return fmt.Errorf("github api: %s", env.Errors[0].Message)
	}
	return json.Unmarshal(env.Data, out)
}

// Fetch gathers everything the cards and README need.
func Fetch(ctx context.Context, c *Client, opt Options) (*Profile, error) {
	var u struct {
		User *struct {
			Login                     string    `json:"login"`
			Name                      string    `json:"name"`
			Bio                       string    `json:"bio"`
			Location                  string    `json:"location"`
			Company                   string    `json:"company"`
			WebsiteURL                string    `json:"websiteUrl"`
			TwitterUsername           string    `json:"twitterUsername"`
			CreatedAt                 time.Time `json:"createdAt"`
			AvatarURL                 string    `json:"avatarUrl"`
			Followers                 count     `json:"followers"`
			Following                 count     `json:"following"`
			PullRequests              count     `json:"pullRequests"`
			Issues                    count     `json:"issues"`
			RepositoriesContributedTo count     `json:"repositoriesContributedTo"`
			PinnedItems               struct {
				Nodes []repoNode `json:"nodes"`
			} `json:"pinnedItems"`
			ContributionsCollection struct {
				ContributionYears    []int `json:"contributionYears"`
				ContributionCalendar struct {
					TotalContributions int `json:"totalContributions"`
					Weeks              []struct {
						ContributionDays []struct {
							Date              string `json:"date"`
							ContributionCount int    `json:"contributionCount"`
							Weekday           int    `json:"weekday"`
						} `json:"contributionDays"`
					} `json:"weeks"`
				} `json:"contributionCalendar"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	}
	if err := c.query(ctx, userQuery, map[string]any{"login": opt.Login}, &u); err != nil {
		return nil, fmt.Errorf("fetch user: %w", err)
	}
	if u.User == nil {
		return nil, fmt.Errorf("user %q not found", opt.Login)
	}
	gu := u.User
	cc := gu.ContributionsCollection

	p := &Profile{
		Login: gu.Login, Name: gu.Name, Bio: strings.TrimSpace(gu.Bio),
		Location: gu.Location, Company: gu.Company, Website: gu.WebsiteURL, Twitter: gu.TwitterUsername,
		AvatarURL: gu.AvatarURL, CreatedAt: gu.CreatedAt,
		Followers: gu.Followers.TotalCount, Following: gu.Following.TotalCount,
		PullRequests: gu.PullRequests.TotalCount, Issues: gu.Issues.TotalCount,
		ContributedTo:         gu.RepositoriesContributedTo.TotalCount,
		ContributionsLastYear: cc.ContributionCalendar.TotalContributions,
	}
	for _, w := range cc.ContributionCalendar.Weeks {
		for _, d := range w.ContributionDays {
			t, err := time.Parse(time.DateOnly, d.Date)
			if err != nil {
				return nil, fmt.Errorf("bad calendar date %q: %w", d.Date, err)
			}
			p.Calendar = append(p.Calendar, Day{Date: t, Count: d.ContributionCount, Weekday: d.Weekday})
		}
	}
	for _, n := range gu.PinnedItems.Nodes {
		p.Pinned = append(p.Pinned, n.toRepo())
	}

	repos, err := c.repos(ctx, opt)
	if err != nil {
		return nil, err
	}
	p.applyRepos(repos, opt)

	if err := c.yearly(ctx, p, opt.Login, cc.ContributionYears); err != nil {
		return nil, err
	}

	if p.AvatarURL != "" {
		// A missing avatar falls back to initials, so don't fail the run over it.
		if uri, err := dataURI(ctx, c.http, p.AvatarURL); err == nil {
			p.Avatar = uri
		}
	}

	p.finalize()
	return p, nil
}

func (c *Client) repos(ctx context.Context, opt Options) ([]repoNode, error) {
	vars := map[string]any{"login": opt.Login, "cursor": nil, "privacy": "PUBLIC"}
	if opt.IncludePrivate {
		vars["privacy"] = nil
	}
	var all []repoNode
	for {
		var r struct {
			User struct {
				Repositories struct {
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
					Nodes []repoNode `json:"nodes"`
				} `json:"repositories"`
			} `json:"user"`
		}
		if err := c.query(ctx, reposQuery, vars, &r); err != nil {
			return nil, fmt.Errorf("fetch repositories: %w", err)
		}
		rs := r.User.Repositories
		all = append(all, rs.Nodes...)
		if !rs.PageInfo.HasNextPage {
			return all, nil
		}
		vars["cursor"] = rs.PageInfo.EndCursor
	}
}

func (p *Profile) applyRepos(repos []repoNode, opt Options) {
	type agg struct {
		color string
		size  int64
	}
	byLang := map[string]*agg{}
	var total int64
	var public []Repo

	for _, r := range repos {
		if !r.IsPrivate {
			p.Repos++
			p.Stars += r.StargazerCount
			p.Forks += r.ForkCount
			public = append(public, r.toRepo())
		}
		if containsFold(opt.ExcludeRepos, r.Name) {
			continue
		}
		if len(r.Languages.Edges) > 0 {
			p.LanguageRepos++
		}
		for _, e := range r.Languages.Edges {
			if containsFold(opt.ExcludeLanguages, e.Node.Name) {
				continue
			}
			a := byLang[e.Node.Name]
			if a == nil {
				a = &agg{color: e.Node.Color}
				byLang[e.Node.Name] = a
			}
			a.size += e.Size
			total += e.Size
		}
	}

	langs := make([]Language, 0, len(byLang))
	for name, a := range byLang {
		color := a.color
		if color == "" {
			color = "#8b949e"
		}
		langs = append(langs, Language{Name: name, Color: color, Bytes: a.size})
	}
	slices.SortFunc(langs, func(a, b Language) int {
		return cmp.Or(cmp.Compare(b.Bytes, a.Bytes), strings.Compare(a.Name, b.Name))
	})
	if n := opt.TopLanguages; n > 0 && len(langs) > n {
		other := Language{Name: "Other", Color: "#8b949e"}
		for _, l := range langs[n-1:] {
			other.Bytes += l.Bytes
		}
		langs = append(langs[:n-1], other)
	}
	for i := range langs {
		if total > 0 {
			langs[i].Percent = float64(langs[i].Bytes) / float64(total) * 100
		}
	}
	p.Languages = langs

	// No pinned repos? Feature the most-starred public ones instead.
	if len(p.Pinned) == 0 {
		slices.SortStableFunc(public, func(a, b Repo) int { return b.Stars - a.Stars })
		p.Pinned = public[:min(len(public), 4)]
	}
}

// yearly sums commits and contributions for every year the account has been active.
// Each contributionsCollection may span at most one year, so years are aliased into one query.
func (c *Client) yearly(ctx context.Context, p *Profile, login string, years []int) error {
	if len(years) == 0 {
		return nil
	}
	var b strings.Builder
	b.WriteString("query($login: String!) { user(login: $login) {")
	for _, y := range years {
		fmt.Fprintf(&b, ` y%d: contributionsCollection(from: "%d-01-01T00:00:00Z", to: "%d-12-31T23:59:59Z") {
			totalCommitContributions contributionCalendar { totalContributions } }`, y, y, y)
	}
	b.WriteString("} }")

	var r struct {
		User map[string]struct {
			TotalCommitContributions int `json:"totalCommitContributions"`
			ContributionCalendar     struct {
				TotalContributions int `json:"totalContributions"`
			} `json:"contributionCalendar"`
		} `json:"user"`
	}
	if err := c.query(ctx, b.String(), map[string]any{"login": login}, &r); err != nil {
		return fmt.Errorf("fetch yearly contributions: %w", err)
	}
	thisYear := fmt.Sprintf("y%d", time.Now().UTC().Year())
	for key, y := range r.User {
		p.Commits += y.TotalCommitContributions
		p.Contributions += y.ContributionCalendar.TotalContributions
		if key == thisYear {
			p.CommitsThisYear = y.TotalCommitContributions
			p.ContributionsThisYear = y.ContributionCalendar.TotalContributions
		}
	}
	return nil
}

func dataURI(ctx context.Context, hc *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("avatar: %s", resp.Status)
	}
	img, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", err
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/") {
		ct = http.DetectContentType(img)
	}
	return "data:" + ct + ";base64," + base64.StdEncoding.EncodeToString(img), nil
}

func containsFold(list []string, s string) bool {
	return slices.ContainsFunc(list, func(v string) bool { return strings.EqualFold(v, s) })
}
