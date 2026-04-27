# adotkaya.github.io

A minimal, dark-mode personal portfolio and blog built with a custom Go static site generator. Zero client-side JavaScript. Zero frameworks. Just Go, HTML, and CSS.

Live at [adotkaya.github.io](https://adotkaya.github.io)

## What This Is

This is not a template built on Next.js, Astro, or Hugo. It is a single Go program (`main.go`) that reads YAML content files and Markdown blog posts, parses HTML templates, and writes static HTML to disk. The entire site builds in milliseconds.

**Core principles:**
- **Build-time over runtime** — Everything is generated ahead of time.
- **Content as data** — All content lives in YAML and Markdown files.
- **Zero JavaScript** — No client-side scripts, no hydration, no bundles.
- **Minimal dependencies** — 3 direct Go modules vs. thousands of npm packages.

## Features

- **Home page** with bio, latest posts, and featured projects
- **Blog** with single-file posts and folder-based composite posts ( Obsidian vault friendly)
- **Resume** page with skills, experience, and education
- **Projects** page with tech stack icons
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

### 3. Updating Other Pages

Edit the corresponding YAML files in `content/`:
- `resume.yaml` — Summary, skills, education
- `experience.yaml` — Work history
- `projects.yaml` — Project cards with tech stack
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

## License

This project is open source. Feel free to fork it and make it your own.
