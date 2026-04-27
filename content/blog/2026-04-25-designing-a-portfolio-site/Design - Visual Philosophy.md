# Design — Visual Philosophy

## Dark Mode by Default

**What:** Near-black background (`#0a0a0a`), soft grays, no pure white.

**Why:** I live in terminals and dark-themed editors. A bright white portfolio feels foreign to my daily environment.

**How:** CSS custom properties make the palette consistent:
```css
:root {
    --bg: #0a0a0a;
    --bg-elevated: #141414;
    --text-primary: #e5e5e5;
    --text-secondary: #a3a3a3;
    --text-muted: #737373;
    --border: #262626;
}
```

**What it improves:**
- Authenticity — the site feels like my workspace.
- Reduced eye strain for visitors who prefer dark themes.
- Easier to maintain one theme than supporting a toggle.

## No Accent Colors, No Animations, No Distractions

**What:** No primary brand color. No hover effects beyond subtle underline and color shifts. No particles, no scroll-jacking, no cookie banners.

**Why:** Minimalism as a forcing function. If an element doesn't serve the content, it doesn't belong.

**How:** The only interactive effects are `opacity` and `color` transitions on links. No JavaScript animations. No `transform` tricks.

**What it improves:**
- Load speed — no animation libraries, no heavy CSS.
- Accessibility — reduced motion is the default.
- Professional tone — the content speaks, not the chrome.
- Time saved — zero hours spent on animation tuning.

## System Font Stack

**What:** `system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, ...`

**Why:** Custom web fonts add HTTP requests and layout shift. System fonts are already on the user's machine.

**How:** Defined once in `body {}`. No `@font-face` declarations.

**What it improves:**
- Zero font loading time.
- No flash of unstyled text (FOUT).
- Native feel — the site matches the OS typography.
- Saves bandwidth for users on slow connections.

## Single CSS File, No Frameworks

**What:** One `static/style.css` (~1,070 lines). No Tailwind, no Bootstrap, no Sass.

**Why:** A portfolio doesn't need a component library. I know exactly what elements exist.

**How:** Hand-written CSS. Mobile breakpoint at `640px`. Print styles for resume.

**What it improves:**
- No build step for CSS.
- Zero unused styles.
- Full control — no fighting framework defaults.
- Browsers parse one small file instantly.

## Mobile-First Responsive

**What:** The site works on all screen sizes. Navigation collapses on small screens.

**Why:** ~60% of web traffic is mobile.

**How:** Flexbox and grid adapt naturally. Single breakpoint.

**What it improves:**
- Professional appearance on all devices.
- No separate mobile site to maintain.

---

*Related: [[Design - Zero JavaScript]], [[Architecture - Custom Go Generator]]*
