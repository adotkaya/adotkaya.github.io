# Design — Zero JavaScript

## What

No `<script>` tags anywhere. The site is pure HTML + CSS.

## Why

JavaScript is powerful but often unnecessary. For a content site, it's overhead.

**Specific reasons:**
- I don't need interactivity beyond links and navigation.
- Every byte of JS is a byte that must be parsed, compiled, and executed.
- No JS means no runtime bugs, no hydration mismatches, no framework updates.

## How

Everything that *could* be done with JS is done at build time instead:

| Task | Typical JS Approach | My Approach |
|------|---------------------|-------------|
| Reading time | Client-side calculation | Build-time word count / 200 wpm |
| Date formatting | `toLocaleDateString()` | Go's `time.Format()` at build |
| Navigation | React Router | Server-rendered `<a href>` |
| Syntax highlighting | Prism.js on load | Chroma at build |
| Mobile menu toggle | JS event listener | CSS-only (future: checkbox hack) |

## What It Improves

- **Performance:** No parse/compile/execute step on load.
- **Security:** No XSS surface from client-side scripts.
- **Reliability:** No broken features if a script fails to load or is blocked.
- **Battery life:** Especially important on mobile devices.
- **Simplicity:** The mental model is "request → HTML → render." Nothing else.

## The Exception That Proves the Rule

The only JavaScript-adjacent feature I considered was client-side search. If the blog grows beyond ~10 posts, I may add a single vanilla JS file that queries a `search.json` generated at build time. Even then, it would be one small script, not a framework.

---

*Related: [[Architecture - Custom Go Generator]], [[Design - Visual Philosophy]]*
