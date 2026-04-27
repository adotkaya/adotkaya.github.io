# Feature — Bookshelf Page

## What

A dedicated `/books/` page listing books with title, author, and reading status badges.

## Why

Reading lists signal intellectual curiosity. A bookshelf page makes the portfolio feel more personal and less like a resume.

**The `/now` page concept extended:** Just as `/now` shows what I'm doing, `/books` shows what I'm learning.

## How

**Data model:**
```yaml
- title: "Designing Data-Intensive Applications"
  author: "Martin Kleppmann"
  status: "finished"
```

**Template:** `templates/books.html` follows the same minimal design as other pages. Uses the existing `.book-list` CSS class.

**Status badges:** `.book-status` CSS rule renders a small pill (reusing the `.tech-tag` pattern):
- `reading` — currently reading
- `finished` — completed
- `to-read` — on the list

**Generator wiring:**
- Added `Status string` to `Book` struct.
- Added render call in `main.go` for `public/books/index.html`.
- Added `/books/` to sitemap and navigation on all 8 existing templates.

## What It Improves

- **Personality:** Shows what I care about learning, not just what I've built.
- **Dynamic feel:** Status badges make the page feel alive and current.
- **Return visits:** Visitors might check back to see what I've finished.
- **Content reuse:** `books.yaml` already existed but was never rendered. This finally surfaces it.

## Design Detail

The status badge uses the same visual language as tech tags on the resume page — muted background, subtle border, capitalized text. Consistency without introducing new patterns.

---

*Related: [[Feature - Content as Data]], [[Design - Visual Philosophy]]*
