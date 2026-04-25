---
title: "Building This Portfolio: From Idea to Go-Powered Static Site"
date: "2026-04-25"
slug: "building-this-portfolio"
---

I have always believed that a developer's portfolio should say something about who they are before a single line of code is read. It is not a resume. It is a statement of taste, values, and craft. This site is mine.

## Why a Static Site Generator in Go?

Most developers reach for Next.js, Astro, or some JavaScript framework when building a personal site. I wanted something different. Something that reflected what I actually care about: simplicity, performance, and minimal dependencies.

Go is fast. The `html/template` package is robust. And generating static files means zero runtime overhead. No client-side hydration. No bundle size anxiety. Just clean HTML and CSS served directly from a CDN.

The entire generator is a single Go program that reads YAML content files, parses templates, and writes HTML to disk. It runs in milliseconds.

## Design Decisions

### Dark Mode by Default

I spend most of my time in terminals and dark-themed editors. A blinding white portfolio feels foreign. The palette here is intentionally restrained: deep near-black backgrounds, soft off-white text, and muted grays for hierarchy. No accent colors. No distractions.

### Minimalism as Constraint

Every element on this site exists because it needs to. There are no animated particles, no scroll-jacking, no cookie banners. The constraint forces clarity. If something does not serve the content, it does not belong.

### Content as Data

All content lives in YAML files. My projects, books, work history, and even this blog post are separate from the presentation layer. This means I can update my experience or add a new project by editing a text file and pushing to GitHub. The site rebuilds automatically.

## What I Learned

Building the generator taught me more about Go's template system than any tutorial. The block and define patterns for layout inheritance are elegant once you understand them. I also gained respect for the sheer amount of complexity modern frontend frameworks abstract away — and questioned whether that complexity is always necessary.

## What's Next

This site will evolve. I plan to write about distributed systems, Go concurrency patterns, and lessons from my transition from .NET to Go. The architecture supports it: new blog posts are just Markdown files. The generator handles the rest.

If you are reading this and considering building your own site from scratch, my advice is simple: start with static files. Understand the foundation. Add complexity only when the problem demands it.

Thanks for visiting.