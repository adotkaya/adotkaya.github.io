---
title: "Designing a Portfolio Site: Decisions, Tradeoffs, and Lessons"
date: "2026-04-25"
slug: "portfolio-design-decisions"
---

A few weeks ago, I sat down to rebuild my personal website. Not because the old one was broken, but because it no longer felt like *mine*. It was a template I had customized until it was unrecognizable, but the underlying assumptions were still someone else's. I wanted something built from first principles.

This post is about those principles — the decisions I made, the tradeoffs I accepted, and what I learned along the way.

## Why Go Instead of Next.js

Most developers reach for Next.js, Astro, or some JavaScript framework when building a personal site. I wanted something different. Something that reflected what I actually care about: simplicity, performance, and minimal dependencies.

Go is fast. The `html/template` package is robust. And generating static files means zero runtime overhead. No client-side hydration. No bundle size anxiety. Just clean HTML and CSS served directly from a CDN.

The entire generator is a single Go program that reads YAML content files, parses templates, and writes HTML to disk. It builds the entire site in milliseconds.

**The tradeoff:** I lose the ecosystem. No React components to npm install. No Tailwind plugins. No pre-built themes. Everything is hand-written. But for a personal portfolio, that is exactly what I want. Every line of code exists because I wrote it, and I know why it is there.

## Dark Mode by Default

I spend most of my time in terminals and dark-themed editors. A blinding white portfolio feels foreign. The palette is intentionally restrained: deep near-black backgrounds, soft off-white text, and muted grays for hierarchy. No accent colors. No distractions.

I never considered a light mode toggle. Maintaining two color schemes means twice the testing, twice the decisions, and a creeping tendency to add more color to differentiate them. One theme, done well, is enough.

## Zero Client-Side JavaScript

The site has no `<script>` tags. None. Every feature that could be done with JavaScript is done at build time instead: reading time calculated from word count, dates formatted by Go's `time` package, navigation rendered as plain HTML links.

This is not a moral stance against JavaScript. It is a constraint that forces clarity. If something cannot be done with static HTML and CSS, I have to ask whether it is necessary. Most of the time, the answer is no.

The performance implications are real: no parse/compile/execute step, no hydration mismatches, no broken features if a script fails to load. But the bigger benefit is mental. The model is simple: request, HTML, render. Nothing else.

## Content as Data

All content lives in YAML and Markdown files. My projects, books, work history, and blog posts are separate from the presentation layer. This means I can update my experience or add a new project by editing a text file and pushing to GitHub. The site rebuilds automatically.

The generator reads these files into Go structs, renders them through templates, and writes the output. Blog posts support frontmatter for titles, dates, and slugs. Draft posts — marked with `draft: true` — are excluded from all renders, so I can iterate on ideas without publishing them.

This separation is not just organizational. It is philosophical. The content should outlive the design. If I redesign the site in two years, the YAML files will still be valid. The Markdown posts will still render. The design is disposable; the content is not.

## SEO and Social Polish

Minimalism does not mean ignoring discoverability. Every page has Open Graph tags, Twitter Cards, JSON-LD structured data, and canonical URLs. The RSS feed auto-discovers itself via `<link rel="alternate">` in every page header.

The favicon and OG image are both SVG — a dark `@` sign on a near-black background. They match the site's palette exactly. When someone shares a link, the preview card looks intentional, not like a default template.

I added these features not because I obsess over analytics, but because I respect the reader. If someone shares my work, the preview should represent it well. If a search engine indexes it, the structure should be clear.

## Security Through Architecture

The deployed site is static HTML, CSS, and SVG. No backend. No database. No API keys. No runtime code.

This is the safest possible architecture for a public website. There is no SQL injection because there is no SQL. There is no XSS from user input because there is no user input. The GitHub Actions deployment uses OIDC instead of long-lived tokens, so there are no secrets to rotate or leak.

The security model is not layered defenses. It is absence of attack surface.

## What I Learned

Building the generator taught me more about Go's template system than any tutorial. The block and define patterns for layout inheritance are elegant once you understand them. I also gained respect for the sheer amount of complexity modern frontend frameworks abstract away — and questioned whether that complexity is always necessary.

The biggest surprise was how little I missed JavaScript. I expected to hit a wall where static files were not enough. That wall never came. Every feature I wanted — RSS, syntax highlighting, responsive navigation — had a build-time solution.

## What Is Next

This site will evolve. I plan to add table of contents for long posts, tag indexes, and a live-reload development server. The architecture supports all of it: new features are just more Go code, new content is just more Markdown.

If you are reading this and considering building your own site from scratch, my advice is simple: start with static files. Understand the foundation. Add complexity only when the problem demands it. Most of the time, it won't.

Thanks for visiting.
