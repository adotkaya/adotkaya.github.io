# Lessons Learned

## Go's Template System Is Elegant

Building the generator taught me more about Go's `html/template` than any tutorial. The block and define patterns for layout inheritance are powerful once understood. Auto-escaping prevents XSS by default — a security feature that's easy to take for granted.

## Static Files Are Surprisingly Sufficient

Modern frontend frameworks abstract away an enormous amount of complexity. Building from scratch made me question whether that complexity is always necessary. For a content site, the answer is often no.

## Constraints Force Clarity

The decision to use **zero JavaScript** and **zero CSS frameworks** meant every feature had to be justified. Could it be done at build time? Could it be done with pure CSS? If not, did I really need it? This discipline produced a cleaner result than unconstrained development would have.

## Content-First Means Less Code

By separating content (YAML/Markdown) from presentation (templates), I spend more time writing and less time debugging layout issues. The generator handles the rest.

## A Portfolio Is Never "Done"

This site will evolve as I learn and build. The architecture supports that: new blog posts are just Markdown files. New pages are just templates and YAML. The foundation is solid enough to grow without rewriting.

## What I'd Do Differently

Nothing major. If I were starting from absolute zero again, I might:
- Add the `/books/` page earlier (the data was there, unused, for multiple commits).
- Consider a Makefile from day one (`make build`, `make serve`).

But these are polish, not regrets. The core decisions hold up.

---

*Related: [[Future Roadmap]], [[Architecture - Custom Go Generator]]*
