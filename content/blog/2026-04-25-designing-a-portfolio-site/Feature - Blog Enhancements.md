# Feature — Blog Enhancements

## RSS / Atom Feed

**What:** Auto-generated Atom feed at `/feed.xml` containing all posts with title, URL, date, and excerpt.

**Why:** RSS is the open web's native subscription protocol. It doesn't require platforms (Twitter, LinkedIn) to distribute content.

**How:** `generateFeed()` in `main.go` writes XML directly. `<link rel="alternate">` in every page `<head>` for auto-discovery.

**What it improves:**
- Readers can subscribe without creating accounts anywhere.
- Platform independence — my content is mine.
- SEO signal — active feeds indicate fresh content to crawlers.

## Human-Readable Dates

**What:** `2026-04-25` → `April 25, 2026` everywhere dates appear.

**Why:** Machine-readable ISO dates are for `datetime` attributes. Humans prefer natural language.

**How:** `formatDate()` in `main.go` using Go's `time.Parse` + `time.Format`. Stored as `DateFormatted` on the `BlogPost` struct.

**What it improves:**
- Warmth — "April 25, 2026" feels human.
- No DD/MM vs. MM/DD confusion.
- Still accessible — `datetime` attribute preserves machine-readability for screen readers and crawlers.

## Previous / Next Post Navigation

**What:** At the bottom of each blog post, links to chronologically adjacent posts.

**Why:** Single-post pages are dead ends without navigation. Readers who finish one post should be invited to read another.

**How:** Added `PrevPost` and `NextPost` pointers to `PageData` during render loop. Conditionally rendered in template footer.

**What it improves:**
- Increased time on site.
- Better internal linking (SEO benefit).
- Reader-friendly — no need to return to the blog list.

## Auto-Heading IDs

**What:** Markdown headings automatically get `id` attributes for direct linking.

**Why:** Readers should be able to link to specific sections of long posts.

**How:** `parser.AutoHeadingIDs` extension in `gomarkdown`.

**What it improves:**
- Deep linking without manual HTML.
- Table of contents can be generated automatically in the future.
- Better anchor navigation.

## Syntax Highlighting with Chroma

**What:** Code blocks in blog posts are highlighted using the Monokai theme.

**Why:** Technical blog posts need readable code. Plain `<pre>` blocks are hard to scan.

**How:** A regex in `main.go` feeds code to `chroma`, which returns inline-styled HTML. No external CSS file needed.

**What it improves:**
- Colors work immediately — styles are inline.
- Monokai fits the dark theme perfectly.
- Language auto-detection falls back gracefully.

---

*Related: [[Feature - Content as Data]], [[Feature - SEO & Social Polish]]*
