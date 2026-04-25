---
title: "Designing a Portfolio Site: A Complete Dev Log"
date: "2026-04-25"
slug: "designing-a-portfolio-site-complete-dev-log"
---

A few weeks ago, I sat down to rebuild my personal website. Not because the old one was broken, but because it no longer felt like *mine*. It was a template I had customized until it was unrecognizable, but the underlying assumptions were still someone else's. I wanted something built from first principles.

This post is the complete story of that rebuild. It covers every major decision — from the programming language to the color palette to the deployment pipeline — with the reasoning behind each choice and the tradeoffs I accepted. It is longer than a typical blog post because I believe the details matter. If you are building your own site, I hope you find something useful here.

## Core Philosophy

> **Minimalism as constraint.** Every element must earn its place. No animations, no JavaScript frameworks, no unnecessary dependencies. The site should feel like an extension of my terminal.

This philosophy is not about aesthetic purity. It is a practical framework for making decisions. When I am unsure whether to add a feature, I ask: does this serve the content, or does it serve itself? If the answer is the latter, it does not belong.

## Part 1: Why I Built a Custom Generator in Go

Most developers reach for Next.js, Astro, or Hugo when building a personal site. I went in the opposite direction: a custom static site generator written in Go.

The generator is a single program, `main.go`, that reads YAML content files and Markdown blog posts, parses HTML templates, and writes static HTML to disk. It builds the entire site in milliseconds.

### Why Go?

Go is what I write professionally now. The portfolio should reflect that. But there are concrete technical reasons too:

- **Build time:** Milliseconds instead of seconds. No bundler, no transpilation, no tree-shaking.
- **Dependencies:** Three direct Go modules versus a thousand npm packages. The dependency surface is negligible.
- **Zero runtime overhead:** The output is HTML and CSS. No client-side hydration. No bundle size anxiety.

### How It Works

The generator uses four key packages:

| Package | Purpose |
|---------|---------|
| `html/template` | Template rendering with automatic escaping |
| `gopkg.in/yaml.v3` | YAML content parsing |
| `gomarkdown/markdown` | Markdown to HTML conversion |
| `alecthomas/chroma/v2` | Syntax highlighting for code blocks |

The build pipeline is trivial:

```
content/*.yaml    ──┐
content/blog/*.md  ─┼──► main.go ──► public/*.html ──► GitHub Pages
templates/*.html  ──┘        ↑
static/*         ────────────┘
```

There is no configuration file for the build tool because there is no build tool. I run `go run main.go` and the site is generated.

### The Tradeoff

I lose the ecosystem. No React components to install. No Tailwind plugins. No pre-built themes. Everything is hand-written. For a personal portfolio, that is exactly what I want. Every line of code exists because I wrote it, and I know why it is there.

## Part 2: Visual Design Decisions

### Dark Mode by Default

I spend most of my time in terminals and dark-themed editors. A bright white portfolio feels foreign. The palette is intentionally restrained:

```css
:root {
    --bg: #0a0a0a;
    --bg-elevated: #141414;
    --text-primary: #e5e5e5;
    --text-secondary: #a3a3a3;
    --text-muted: #737373;
    --border: #262626;
}
```

I never considered a light mode toggle. Maintaining two color schemes means twice the testing, twice the decisions, and a creeping tendency to add more color to differentiate them. One theme, done well, is enough.

### No Accent Colors, No Animations, No Distractions

There is no primary brand color. The only interactive effects are subtle opacity and color transitions on links. No particles. No scroll-jacking. No cookie banners.

This is minimalism as a forcing function. If an element does not serve the content, it does not belong. The constraint forces clarity. It also saves time: I spent zero hours on animation tuning.

### System Font Stack

I use `system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, ...` instead of custom web fonts. System fonts are already on the user's machine. This eliminates loading time, prevents layout shift, and gives the site a native feel.

### Single CSS File, No Frameworks

The entire site is styled by one `static/style.css` file of roughly 1,070 lines. No Tailwind, no Bootstrap, no Sass.

A portfolio does not need a component library. I know exactly what elements exist. Hand-written CSS means zero unused styles and no fighting framework defaults.

### Mobile-First Responsive

The site works on all screen sizes with a single breakpoint at `640px`. Navigation collapses on small screens. The resume page has dedicated print styles that hide the nav and footer, remove backgrounds, and prevent page breaks inside job entries.

## Part 3: Zero Client-Side JavaScript

The site has no `<script>` tags. None.

This is not a moral stance against JavaScript. It is a constraint that forces clarity. If something cannot be done with static HTML and CSS, I have to ask whether it is necessary. Most of the time, the answer is no.

Here is how common tasks are handled without JavaScript:

| Task | Typical JS Approach | My Approach |
|------|---------------------|-------------|
| Reading time | Client-side calculation | Build-time word count divided by 200 wpm |
| Date formatting | `toLocaleDateString()` | Go's `time.Format()` at build time |
| Navigation | React Router | Server-rendered `<a href>` elements |
| Syntax highlighting | Prism.js on load | Chroma at build time |
| Mobile menu | JS event listener | CSS-only (hamburger with CSS toggle) |

The performance implications are real: no parse/compile/execute step on load, no hydration mismatches, no broken features if a script fails to load. But the bigger benefit is mental. The model is simple: request, HTML, render. Nothing else.

## Part 4: Content as Data

All content lives in YAML and Markdown files. Templates are pure presentation. The two never mix.

### Structured Data in YAML

- `config.yaml` — name, bio, social links
- `resume.yaml` — summary, skills, education
- `experience.yaml` — work history
- `projects.yaml` — project list with tech stacks
- `books.yaml` — reading list with status
- `now.yaml` — current activities
- `uses.yaml` — tools and workflow

### Long-Form Content in Markdown

Blog posts live in `content/blog/*.md` with YAML frontmatter for titles, dates, and slugs. Draft posts — marked with `draft: true` — are excluded from all renders, so I can iterate on ideas without publishing them.

### Why This Matters

The content should outlive the design. If I redesign the site in two years, the YAML files will still be valid. The Markdown posts will still render. The design is disposable; the content is not.

This separation also reduces writing friction. Add a Markdown file, run `go run main.go`, push. No HTML to edit. No layout to worry about.

## Part 5: Blog Enhancements

### RSS / Atom Feed

The site generates an Atom feed at `/feed.xml` containing all posts with title, URL, date, and excerpt. Every page includes `<link rel="alternate">` in the `<head>` for auto-discovery.

RSS is the open web's native subscription protocol. It does not require platforms like Twitter or LinkedIn to distribute content. Readers can subscribe without creating accounts anywhere.

### Human-Readable Dates

Machine-readable ISO dates like `2026-04-25` are used in `datetime` attributes for crawlers and screen readers. Humans see `April 25, 2026`. This is handled by Go's `time.Parse` and `time.Format` at build time.

### Previous / Next Post Navigation

At the bottom of each blog post, links to chronologically adjacent posts invite the reader to continue. This is handled during the render loop by passing `PrevPost` and `NextPost` pointers to the template.

### Syntax Highlighting with Chroma

Code blocks are highlighted using the Monokai theme. A regex in `main.go` extracts code from `<pre><code>` tags, feeds it to Chroma, and inserts inline-styled HTML. No external CSS file is needed for highlighting.

### Auto-Heading IDs

Markdown headings automatically get `id` attributes via `gomarkdown`'s `AutoHeadingIDs` extension. This enables deep linking to specific sections without manual HTML.

## Part 6: Discoverability and Polish

A minimal site does not mean ignoring discoverability. Every page has a comprehensive set of meta tags and structured data.

### Open Graph and Twitter Cards

Every page includes `og:title`, `og:description`, `og:type`, `og:url`, and `og:image`. Blog posts use `og:type = article` and include `article:published_time`. Twitter Card meta tags mirror the Open Graph data.

### JSON-LD Structured Data

- **Home page:** `Person` schema with name, URL, `sameAs` links to GitHub and LinkedIn, job title, and description.
- **Blog posts:** `BlogPosting` schema with headline, author, `datePublished`, URL, and description.
- **All other pages:** `WebPage` schema with name and URL.

### Canonical URLs

Every page includes `<link rel="canonical">` to prevent duplicate content penalties from search engines.

### Sitemap

An auto-generated `sitemap.xml` includes all pages and posts with priorities and change frequencies.

### Favicon and OG Image

Both are SVG files: a dark `@` sign on a near-black background. They match the site's palette exactly. Browser tabs are instantly recognizable, and social shares look intentional rather than like default templates.

## Part 7: The Bookshelf Page

A dedicated `/books/` page lists books with title, author, and reading status badges: `reading`, `finished`, or `to-read`.

Reading lists signal intellectual curiosity. Status badges make the page feel alive and current. The design reuses the existing `.book-list` CSS class and the `.tech-tag` visual pattern for badges, maintaining consistency without introducing new patterns.

The page was added after the initial build, which revealed a useful pattern: `books.yaml` had existed for weeks but was never rendered. Separating content from presentation means data can exist before its UI is designed, with no breakage.

## Part 8: Security Through Architecture

The deployed site is static HTML, CSS, and SVG. No backend. No database. No API keys. No runtime code.

This is the safest possible architecture for a public website. There is no SQL injection because there is no SQL. There is no XSS from user input because there is no user input. The GitHub Actions deployment uses OIDC (`id-token: write`) instead of long-lived tokens, so there are no secrets to rotate or leak.

A security scan of the repository found no credentials, no tokens, and no secrets in the git history. The `public/` directory is gitignored and never committed. The only PII is the content intentionally displayed on the portfolio itself.

The security model is not layered defenses. It is absence of attack surface.

## Part 9: Deployment

The site deploys automatically via GitHub Actions on every push to `main`.

### Workflow

```
Push to main ──► Checkout ──► Setup Go ──► go run main.go ──► Upload artifact ──► Deploy to Pages
```

### Key Details

- **Trigger:** Push to `main` or manual `workflow_dispatch`.
- **Authentication:** OIDC. No personal access tokens.
- **Permissions:** Minimal — `contents: read`, `pages: write`, `id-token: write`.
- **Cost:** $0. GitHub Pages is free for public repositories.

The entire pipeline completes in under a minute. The site updates roughly 30 seconds after I push.

## Part 10: Lessons Learned

### Go's Template System Is Elegant

Building the generator taught me more about Go's `html/template` package than any tutorial. The block and define patterns for layout inheritance are powerful once understood. Auto-escaping prevents XSS by default — a security feature that is easy to take for granted.

### Static Files Are Surprisingly Sufficient

Modern frontend frameworks abstract away an enormous amount of complexity. Building from scratch made me question whether that complexity is always necessary. For a content site, the answer is often no.

### Constraints Force Clarity

The decision to use zero JavaScript and zero CSS frameworks meant every feature had to be justified. Could it be done at build time? Could it be done with pure CSS? If not, did I really need it? This discipline produced a cleaner result than unconstrained development would have.

### A Portfolio Is Never "Done"

This site will evolve as I learn and build. The architecture supports that: new blog posts are just Markdown files. New pages are just templates and YAML. The foundation is solid enough to grow without rewriting.

## Part 11: Future Roadmap

### High Priority

- **Table of Contents for blog posts:** Auto-generated from Markdown headings. Pure HTML, zero JS.
- **Tags and tag index pages:** `tags: [go, architecture]` in frontmatter, with generated `/blog/tag/go/` pages.
- **Tools section on home page:** `tools.yaml` already exists with 11 tools. A visual grid would complete the homepage.

### Medium Priority

- **Live reload dev server:** A file-watcher loop in Go that rebuilds on changes. Essential for writing sessions.
- **Client-side search:** `main.go` generates `search.json`; a single vanilla JS file powers a `/search/` page. Valuable once the blog exceeds ~10 posts.

### Low Priority / Polish

- Accessibility improvements: skip navigation links, `focus-visible` styles.
- HTML minification to strip whitespace from generated output.
- Self-hosting SimpleIcons to remove the external CDN dependency.

## Conclusion

If you are considering building your own site from scratch, my advice is simple: start with static files. Understand the foundation. Add complexity only when the problem demands it. Most of the time, it won't.

The best portfolio is not the one with the most features. It is the one that most accurately represents who you are. This one is mine.

Thanks for reading.
