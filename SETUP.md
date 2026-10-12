# Profile README generator

A small Go app (stdlib only) that pulls your stats from the GitHub GraphQL API and renders
animated SVG cards in dark and light versions, plus `README.md` from a template. A GitHub Actions
workflow re-runs it every 6 hours and commits the result.

```
main.go                     entrypoint
internal/stats/             GraphQL queries, language aggregation, streaks
internal/card/              SVG cards (profile, languages, activity)
internal/readme/            README template rendering
templates/README.md.tmpl    the README layout — edit freely
config.json                 your settings
assets/                     generated SVGs (committed by the workflow)
```

## Setup

1. Push this folder to a **public** repo named exactly like your username (`Demoniskk/Demoniskk`).
   GitHub shows that repo's README on your profile.
2. In the repo, open **Settings → Actions → General → Workflow permissions** and choose
   **Read and write permissions**.
3. *(Optional but recommended)* Create a fine-grained or classic PAT (classic: `read:user` + `repo`)
   and save it as the repo secret **`GH_STATS_TOKEN`**. Without it the built-in token is used, which
   only counts public activity.
4. Run the workflow once from the **Actions** tab (**Update profile README → Run workflow**).

## Local preview

```sh
go run . -demo                        # sample data, no token
GH_TOKEN=$(gh auth token) go run .    # your real stats
```

## Customising

- `config.json`: accent gradient, status pill, tagline, about lines, skill icons
  ([skillicons.dev ids](https://github.com/tandpfun/skill-icons#icons-list)), links, excluded languages/repos.
- `templates/README.md.tmpl`: layout. `{{ card "profile" "alt" }}` inserts a theme-aware card.
  `.User` holds all stats (see `internal/stats/profile.go`).
