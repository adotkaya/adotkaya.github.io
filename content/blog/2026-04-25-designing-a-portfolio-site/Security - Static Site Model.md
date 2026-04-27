# Security — Static Site Model

## What

The deployed site is static HTML + CSS + SVG. No backend, no database, no API keys, no runtime code.

## Why

Every additional service is an attack surface and a maintenance burden. For a personal portfolio, this attack surface should be zero.

## How

**No secrets in repository:**
- Scanned git history: no API keys, tokens, passwords, or credentials committed.
- No `.env` files. No config files with hidden values.

**No runtime:**
- Go generator runs at build time only.
- The deployed artifact is HTML/CSS/SVG.
- Nothing executes on the server or in the browser.

**GitHub Actions security:**
- Uses OIDC (`id-token: write`) for Pages deployment.
- No long-lived personal access tokens stored as secrets.
- Permissions are minimal: `contents: read`, `pages: write`.

## What It Improves

- **Nothing to hack:** No SQL injection, no XSS from user input, no auth bypass.
- **Nothing to leak:** No API keys, no database credentials.
- **Nothing to patch:** No runtime dependencies to update for security advisories.
- **No costs:** GitHub Pages is free and scales infinitely for static content.

## Minor Considerations

| Risk | Mitigation |
|------|------------|
| External CDN (SimpleIcons) | Icons load as `<img>`, not `<script>`. XSS risk is minimal. |
| Email in git commits | Optional: enable GitHub's "Keep my email addresses private" setting. |
| Intentional PII exposure | Full name, location, work history are by design for a public portfolio. |

## Verdict

**Zero meaningful security risk.** This is the safest possible architecture for a public website.

---

*Related: [[Architecture - Custom Go Generator]], [[Deployment - GitHub Actions]]*
