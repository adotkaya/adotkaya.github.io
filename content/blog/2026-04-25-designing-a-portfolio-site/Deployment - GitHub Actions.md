# Deployment — GitHub Actions

## What

A GitHub Actions workflow that builds the site on every push to `main` and deploys to GitHub Pages automatically.

## Why

Push-to-deploy is the ideal developer experience. I write, commit, push, and the site updates. No manual FTP, no server management, no deployment scripts on my machine.

## How

**Workflow file:** `.github/workflows/deploy.yml`

**Pipeline:**
```
Push to main ──► Checkout ──► Setup Go ──► go run main.go ──► Upload artifact ──► Deploy to Pages
```

**Key details:**
- **Trigger:** Push to `main` or manual `workflow_dispatch`.
- **Go version:** 1.21 (workflow) — note: `go.mod` specifies 1.22, minor drift.
- **Authentication:** OIDC (`id-token: write`). No PATs, no secrets.
- **Permissions:** Minimal — `contents: read`, `pages: write`, `id-token: write`.

## What It Improves

- **Zero manual deployment:** The site updates in ~30 seconds after push.
- **Auditable:** Build and deploy history is visible in GitHub's UI.
- **Secure:** OIDC is the modern standard. No long-lived tokens to rotate.
- **Free:** GitHub Pages costs $0 for public repos.

## The Build Step

```yaml
- name: Build site
  run: |
    go mod download
    go run main.go
```

This downloads the 3 dependencies and generates the entire site in milliseconds.

## Future Improvement

Update `go-version: '1.21'` → `'1.22'` in the workflow to match `go.mod`. Not urgent, but avoids confusion.

---

*Related: [[Security - Static Site Model]], [[Architecture - Custom Go Generator]]*
