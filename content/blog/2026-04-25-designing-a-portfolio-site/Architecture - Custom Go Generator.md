# Architecture — Custom Go Generator

## What

A single Go program (`main.go`) reads YAML content files and Markdown blog posts, parses HTML templates, and writes static HTML to disk.

## Why Not Next.js / Astro / Hugo?

Most developers default to JavaScript frameworks. I wanted the site to reflect my actual values — simplicity, performance, minimal dependencies. Go is what I write professionally now. The portfolio itself should demonstrate that.

**Specific reasons:**
- **Build time:** Milliseconds instead of seconds.
- **Dependency surface:** 3 direct Go modules vs. 1,000+ npm packages.
- **Zero runtime overhead:** No client-side hydration, no bundle size anxiety.
- **Learning value:** Deeper understanding of template systems and static site fundamentals.

## How It Works

```
content/*.yaml  ──┐
content/blog/*.md ─┼──► main.go ──► public/*.html ──► GitHub Pages
templates/*.html ──┘         ↑
static/* ────────────────────┘
```

**Key packages:**
| Package | Purpose |
|---------|---------|
| `html/template` | Template rendering |
| `gopkg.in/yaml.v3` | YAML content parsing |
| `gomarkdown/markdown` | Markdown → HTML |
| `alecthomas/chroma/v2` | Syntax highlighting |

## What It Improves

- **Interview value:** The generator itself is a project I can discuss. It shows systems thinking, not just framework usage.
- **Maintenance:** No version conflicts, no deprecated packages, no security advisories for frontend tooling.
- **Speed:** The entire site builds in under a second.

## Build Pipeline

```bash
go run main.go    # Generates public/
# Commit & push  # GitHub Actions deploys automatically
```

**No build tool.** No `webpack.config.js`. No `vite.config.ts`. Just Go.

---

*Related: [[Design - Zero JavaScript]], [[Security - Static Site Model]]*
