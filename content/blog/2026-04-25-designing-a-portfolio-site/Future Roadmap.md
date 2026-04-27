# Future Roadmap

## High Priority

### Table of Contents for Blog Posts
Auto-generate from Markdown headings (`h2`, `h3`). Injected at the top of long posts. Pure HTML, zero JS. Becomes valuable once posts exceed 1,000 words.

### Tags & Tag Index Pages
`tags: [go, architecture]` in frontmatter. Generate `/blog/tag/go/` index pages. Valuable once there are 5+ posts.

### Tools Section on Home Page
`tools.yaml` exists with 11 tools + SimpleIcons. A visual grid below Projects would add color and context to the homepage.

## Medium Priority

### Live Reload Dev Server
A simple file-watcher loop in Go that rebuilds on `content/` or `templates/` changes. Huge quality-of-life improvement for writing sessions.

### Client-Side Search
`main.go` generates `public/search.json` with post titles, excerpts, tags. A single vanilla JS file (`static/search.js`) powers a `/search/` page. No external service. Valuable once blog exceeds ~10 posts.

### Makefile
Standard conventions: `make build`, `make serve`, `make clean`.

## Low Priority / Polish

### Accessibility Improvements
- Skip navigation link ("Skip to main content") for keyboard users.
- `focus-visible` styles for clearer keyboard navigation.
- Normalize `<time>` element usage across all pages.

### HTML Minification
Strip whitespace from generated HTML. Saves ~10-20% file size. Marginal gain for a small site.

### Self-Host SimpleIcons
Download tech stack icons into `static/` to remove external CDN dependency entirely.

### Content Security Policy
Add `<meta http-equiv="Content-Security-Policy" ...>` for extra XSS hardening. Overkill for a static site with no scripts, but good hygiene.

---

*Related: [[Lessons Learned]], [[Feature - Blog Enhancements]]*
