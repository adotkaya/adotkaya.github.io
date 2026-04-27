# Feature — Content as Data

## What

All content lives in `content/*.yaml` and `content/blog/*.md`. Templates are pure presentation. The two never mix.

## Why

Separation of concerns. I can update my experience or add a blog post without touching HTML or CSS.

**The alternative:** Hardcoding content in templates. Every edit requires opening HTML files, finding the right tags, not breaking structure. That's friction.

## How

**YAML files for structured data:**
- `config.yaml` — name, bio, social links
- `resume.yaml` — summary, skills, education
- `experience.yaml` — work history
- `projects.yaml` — project list with tech stacks
- `books.yaml` — reading list with status
- `now.yaml` — current activities
- `uses.yaml` — tools and workflow

**Markdown for long-form:**
- `content/blog/*.md` with YAML frontmatter

**Build process:**
```go
data.Config = mustParseYAML[Config]("content/config.yaml")
data.Posts = loadBlogPosts()
renderTemplate("templates/index.html", "public/index.html", data)
```

## What It Improves

- **Writing friction is near zero:** Add a Markdown file, run `go run main.go`, push.
- **Content is version-controlled:** No separate CMS database.
- **Portable:** YAML is universal. I'm not locked into any platform.
- **Enables draft support:** `draft: true` in frontmatter excludes a post from all renders.

## Draft Post Support

**What:** `draft: true` in frontmatter excludes the post from home, blog list, RSS, and prevents individual page generation.

**Why:** I need a place to iterate on posts without publishing them.

**How:** Added `Draft bool` to frontmatter struct. Early return if draft.

**What it improves:**
- I can commit work-in-progress posts to git without them going live.
- No risk of accidentally publishing unfinished thoughts.
- Draft files live alongside published ones.

---

*Related: [[Feature - Blog Enhancements]], [[Architecture - Custom Go Generator]]*
