# adotkaya.github.io

A minimal, dark-mode personal portfolio and blog built with a custom Go static site generator. Zero client-side JavaScript. Zero frameworks. Just Go, HTML, and CSS.

Live at [adotkaya.github.io](https://adotkaya.github.io)
If you want to build your own version, you can clone this repository [port-gen](https://github.com/adotkaya/port-gen)

## What This Is

This is not a template built on Next.js, Astro, or Hugo. It is a single Go program (`main.go`) that reads YAML content files and Markdown blog posts, parses HTML templates, and writes static HTML to disk. The entire site builds in milliseconds.

**Core principles:**
- **Build-time over runtime** — Everything is generated ahead of time.
- **Content as data** — All content lives in YAML and Markdown files.
- **Zero JavaScript** — No client-side scripts, no hydration, no bundles.
- **Minimal dependencies** — 3 direct Go modules vs. thousands of npm packages.

## Features

- **Home page** with bio, latest posts, and featured projects (auto-fetched from GitHub)
- **Blog** with single-file posts and folder-based composite posts (Obsidian vault friendly)
- **Resume** page with skills, experience, and education — shows only pinned projects
- **Projects** page auto-fetched from GitHub with language detection, fork/archived badges, and YAML overrides
- **CSS-only mobile nav** — hamburger menu without any JavaScript
- **Now** page for current status
- **Uses** page for tools and gear
- **Bookshelf** page with reading status badges
- **RSS/Atom feed** (`/feed.xml`)
- **Sitemap** (`/sitemap.xml`) and `robots.txt`
- **SEO** — Open Graph, Twitter Cards, JSON-LD structured data, canonical URLs
- **Syntax highlighting** for code blocks via Chroma
- **Dark mode by default** with a single hand-written CSS file
- **GitHub Actions** deployment to GitHub Pages via OIDC

## Quick Start

### Prerequisites

- [Go](https://go.dev/dl/) 1.22 or later
- Git

### Clone and Run

```bash
git clone https://github.com/adotkaya/adotkaya.github.io.git
cd adotkaya.github.io

# Download dependencies and build the site
go mod download
go run main.go

# Serve locally (Python 3)
cd public && python -m http.server 8000
# Or with Node.js: npx serve .
```

Open `http://localhost:8000` in your browser.

## Project Structure

```
.
├── main.go              # The static site generator
├── go.mod               # Go module definition
├── content/             # Your content (YAML + Markdown)
│   ├── config.yaml      # Site metadata, social links, bio
│   ├── pages.yaml       # Page subtitles and nav links
│   ├── resume.yaml      # Resume data
│   ├── experience.yaml  # Work experience
│   ├── projects.yaml    # Project listings
│   ├── now.yaml         # /now page content
│   ├── uses.yaml        # /uses page content
│   ├── books.yaml       # /books page content
│   ├── tools.yaml       # Toolbox icons
│   └── blog/            # Blog posts
│       ├── hello-world.md
│       └── 2026-04-27-my-post/   # Folder-based post
│           ├── index.md          # Post metadata
│           └── 01-introduction.md
├── content-example/     # Example content files (reference)
├── templates/           # HTML templates (Go html/template)
│   ├── index.html
│   ├── blog-post.html
│   ├── resume.html
│   └── ...
├── static/              # Static assets (CSS, images, favicon)
│   ├── style.css
│   └── ...
├── public/              # Generated site output (gitignored)
└── .github/workflows/   # GitHub Actions deployment
```

## Customizing Content

All personal data lives in `content/*.yaml`. There is a matching `content-example/` directory with annotated templates showing the expected structure.

### 1. Site Config (`content/config.yaml`)

```yaml
name: "Your Name"
username: "yourusername"
bio: "Your tagline"
github: "https://github.com/yourusername"
linkedin: "https://linkedin.com/in/yourusername"
location: "City, Country"
about: "A short paragraph about yourself."
```

### 2. Adding a Blog Post

**Single-file post:**

Create `content/blog/my-post.md`:

```markdown
---
title: "My Post Title"
date: "2026-04-27"
slug: "my-post"
---

Your content here in Markdown.
```

**Folder-based post** (great for longform or Obsidian vaults):

Create `content/blog/2026-04-27-my-post/` with an `index.md`:

```markdown
---
title: "My Longform Post"
date: "2026-04-27"
slug: "my-post"
---
```

Then add any number of `.md` files in the same folder. They are concatenated alphabetically:

```
content/blog/2026-04-27-my-post/
├── index.md
├── 01-introduction.md
├── 02-deep-dive.md
└── 03-conclusion.md
```

- Each sub-file can have its own frontmatter (it will be stripped).
- Obsidian wiki-links like `[[Another Note]]` are automatically converted to plain text.
- Draft posts or folders prefixed with `_` are skipped.

### 3. Projects — Auto-Fetched from GitHub + YAML Override

Your projects page is populated automatically at build time by fetching your public GitHub repos via the GitHub API. `content/projects.yaml` acts as an optional overlay to customize descriptions, tech stacks, pinning, and exclusions.

**How it works:**
- On every build, `main.go` calls `https://api.github.com/users/{username}/repos`
- All public repos are imported with auto-detected language icons
- YAML entries override GitHub data for repos with matching names
- Repos are sorted by last push date (newest first)
- Forks and archived repos are included with small badges

**YAML fields (all optional):**

```yaml
- name: "my-repo"           # must match GitHub repo name
  description: "Custom description overrides GitHub's"
  url: "https://github.com/you/my-repo"
  tech_stack:               # overrides auto-detected language
    - "go"
    - "postgresql"
  pinned: true              # shows on home page and resume (max 6)
  exclude: true             # hides repo everywhere
```

**Examples:**

Pin a repo to the home page:
```yaml
- name: "cool-project"
  pinned: true
```

Hide a repo you don't want listed:
```yaml
- name: "old-fork"
  exclude: true
```

Add a private repo or external project not on GitHub:
```yaml
- name: "internal-tool"
  description: "A private project at work"
  url: "https://gitlab.com/you/internal-tool"
  tech_stack: ["go", "redis"]
  pinned: true
```

**Resume page:** The resume page shows **only pinned projects** (no fallback). If you have no pinned repos, the Projects section on the resume is hidden entirely. This keeps your resume concise and curated.

**Rate limiting:** The GitHub API allows 60 requests/hour without authentication (we make 1 request per build). This is plenty for normal development. If you hit the limit during heavy development, you can optionally set a `GITHUB_TOKEN`:

```bash
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
```

This is completely optional. The site builds fine without it.

**Build-time vs. runtime:** The GitHub API is called **only when the site is built** (`go run main.go`), not when a visitor loads the page. This means:

- Creating a new repo on GitHub does **not** instantly update the site.
- The new repo appears **after the next deploy** (the next time `go run main.go` runs).
- The site is static HTML — there is no server, no database, and no background job checking for updates.

**Deployment flow:**

```
You create a new repo on GitHub
        ↓
You push code to main (or trigger deploy.yml manually)
        ↓
GitHub Actions runs: go run main.go
        ↓
main.go fetches all repos from the GitHub API
        ↓
Static HTML is generated with the new repo included
        ↓
public/ folder is deployed to GitHub Pages
        ↓
Site now shows the new repo
```

To force a rebuild without pushing code, go to **Actions → Deploy to GitHub Pages → Run workflow**.

**Daily auto-rebuild:** The workflow also runs on a schedule (`0 3 * * *` — every day at 3 AM UTC). This keeps your projects list in sync with GitHub automatically, even if you don't push any code. The scheduled rebuild is safe and public: it uses your repository's minimal permissions, consumes only 1 API request per day, and has no access to secrets.

### 4. Updating Other Pages

Edit the corresponding YAML files in `content/`:
- `resume.yaml` — Summary, skills, education
- `experience.yaml` — Work history
- `now.yaml` — What you're doing now
- `uses.yaml` — Tools and gear
- `books.yaml` — Reading list with status badges
- `pages.yaml` — Navigation subtitles and links
- `tools.yaml` — SimpleIcons slugs for the toolbox section

## Customizing Design

The entire design is controlled by one file: `static/style.css`.

- **Colors** — CSS custom properties at the top of the file
- **Typography** — System font stack (no web fonts)
- **Layout** — Mobile-first with a single breakpoint at `640px`
- **Print styles** — Included for the resume page

Templates are standard Go `html/template` files in `templates/`. No special framework syntax — just Go templates with `{{ .Variable }}` syntax.

## Deployment

This repo includes a GitHub Actions workflow (`.github/workflows/deploy.yml`) that:

1. Builds the site on every push to `main`
2. Deploys the `public/` directory to GitHub Pages automatically

To use it:

1. Fork this repo
2. Go to **Settings → Pages** and set the source to **GitHub Actions**
3. Push your changes to the `main` branch

That's it. No tokens, no secrets. The workflow uses OIDC for secure deployment.

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/gomarkdown/markdown` | Markdown → HTML |
| `github.com/alecthomas/chroma/v2` | Syntax highlighting |
| `gopkg.in/yaml.v3` | YAML parsing |

## Why Go?

- **Build time:** Milliseconds instead of seconds.
- **Dependency surface:** 3 Go modules vs. 1,000+ npm packages.
- **Zero runtime overhead:** No client-side hydration, no bundle size anxiety.
- **Learning value:** Deeper understanding of template systems and static site fundamentals.
- **Interview value:** The generator itself demonstrates systems thinking.

## Security

This project is designed to be secure by default:

- **Zero runtime attack surface** — The output is plain static HTML/CSS. No server, no database, no APIs to exploit.
- **No secrets in the repository** — No API keys, tokens, or credentials are hardcoded. The optional `GITHUB_TOKEN` is read from environment variables only.
- **No client-side JavaScript** — No scripts, no cookies, no tracking, no XSS vectors.
- **OIDC deployment** — GitHub Actions uses OpenID Connect to deploy to Pages. No long-lived personal access tokens stored as secrets.
- **Go template auto-escaping** — All variables are HTML-escaped by default. User-authored Markdown is the only unescaped content.

## License

This project is open source. Feel free to fork it and make it your own.
