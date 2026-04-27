# Feature — SEO & Social Polish

## What

A comprehensive set of meta tags and structured data to make the site discoverable and shareable.

## Why

Without these, sharing a link on Twitter/LinkedIn/Discord produces a bare URL. Search engines struggle to understand page structure and authorship.

## How

### Open Graph Tags (Every Page)
```html
<meta property="og:title" content="...">
<meta property="og:description" content="...">
<meta property="og:type" content="website|article">
<meta property="og:url" content="...">
<meta property="og:image" content=".../og-image.svg">
```

### Twitter Cards (Every Page)
```html
<meta name="twitter:card" content="summary">
<meta name="twitter:title" content="...">
<meta name="twitter:description" content="...">
```

### JSON-LD Structured Data
- **Home:** `Person` schema (name, url, sameAs links to GitHub/LinkedIn, jobTitle)
- **Blog Posts:** `BlogPosting` schema (headline, author, datePublished, url, description)
- **Other Pages:** `WebPage` schema (name, url)

### Canonical URLs
```html
<link rel="canonical" href="https://adotkaya.github.io/.../">
```

### Sitemap
Auto-generated `sitemap.xml` with priorities and change frequencies for every page and post.

### Favicon & OG Image
- `favicon.svg` — `@` glyph on dark background
- `og-image.svg` — 1200×630 dark `@` mark for social previews

## What It Improves

- **Social shares:** Rich previews with title, description, and image.
- **Search rankings:** Search engines understand who I am, what I write, and how pages relate.
- **Click-through rate:** Structured data can produce rich snippets in search results.
- **SEO hygiene:** Canonical URLs prevent duplicate content penalties.

## Page-Type Specifics

| Page | `og:type` | JSON-LD Schema |
|------|-----------|----------------|
| Home | `website` | `Person` |
| Blog Post | `article` | `BlogPosting` |
| All Others | `website` | `WebPage` |

---

*Related: [[Feature - Blog Enhancements]], [[Design - Visual Philosophy]]*
