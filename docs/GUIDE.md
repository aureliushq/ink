# Building Ink, a Static Site Generator in Go — A Builder's Guide

A roadmap for building a small, focused SSG you actually understand, end to end. No code — just the concepts, the architecture, the packages worth reaching for, and the order to build things in.

---

## 1. What a static site generator is

A static site generator turns **content + templates + assets** into a folder of plain HTML/CSS/JS files **at build time**. There's no server doing work per request. The output is just files; any static host or CDN can serve them.

It helps to place SSGs against the alternatives:

- **Dynamic sites** (Rails, Django, a Go web app): the server renders HTML on *every request*. Flexible, but needs a running process and usually a database.
- **SPAs** (React/Vue): the browser renders everything *client-side* from JS. Great for app-like interactivity, worse for content-heavy sites and first-paint.
- **SSGs**: render *once*, at build time. The reader gets pre-baked HTML.

Why people pick SSGs for blogs, docs, and project sites: they're fast (CDN-served static files), cheap-to-free to host, secure (no runtime, no DB, tiny attack surface), and the whole site lives in Git. The tradeoff is that anything truly dynamic (comments, search, auth) has to be handled client-side or by an external service.

The goal here — something smaller and simpler than Hugo, scoped to a handful of projects — is exactly the sweet spot where rolling your own makes sense. Hugo is enormous because it supports everyone's use case. A purpose-built tool only has to support yours.

---

## 2. How they work (the pipeline)

Every SSG, Hugo included, is fundamentally the same pipeline. Internalize this and the rest is detail:

1. **Discover** — walk the content directory, find all the source files (markdown).
2. **Parse** — split each file into *frontmatter* (metadata: title, date, tags) and *body* (markdown). Convert the body to HTML.
3. **Model** — assemble an in-memory representation of the site: a list of pages, each with its metadata, rendered HTML, and target output path. Derive collections (all posts, posts by tag, etc.).
4. **Render** — for each page, pick a template/layout and execute it, injecting the page's HTML and metadata. This produces the final HTML string.
5. **Assets** — copy (and optionally transform) static files: CSS, images, fonts.
6. **Write** — emit everything to an output directory (`public/`, `dist/`).
7. **(Dev only)** — optionally run a local server that watches for changes, rebuilds, and live-reloads the browser.

The whole thing is a function: `source tree → output tree`. Keep that mental model front and center; it keeps the design honest.

---

## 3. Architecture for your SSG

Separate the **engine** (a library) from the **CLI** (a thin binary that calls the library). This pays off later when you want to test the engine directly, embed it, or wrap it in a GitHub Action.

A package layout that scales well without being over-engineered:

```
yourssg/
  cmd/yourssg/        # main package: CLI entrypoint, flag/arg parsing
  internal/
    config/           # load + validate site config
    content/          # discover files, parse frontmatter + markdown, build page model
    render/           # load templates, execute them against the page model
    assets/           # copy/transform static files
    build/            # the orchestrator: wires the pipeline together
    server/           # dev server: file watching + live reload
  themes/             # built-in themes, all embedded via go:embed themes/*
    default/          #   each is a full set of reserved templates + its own CSS
    minimal/          #   theme.name in config picks which one feeds the lookup
  go.mod
```

The engine ships **one or more built-in themes** (all embedded with `go:embed`), a config key selects which one, and a site can still override individual templates on top — see the design principle below.

And the **convention for a site that your tool builds** (this is the contract your users — including you — follow):

```
my-site/
  config.toml         # site-level settings
  content/
    index.md
    posts/
      hello-world.md
  layouts/            # OPTIONAL — override any built-in template by name; omit to use built-ins
    base.html         # the shell (html/head/body)
    single.html       # a single post/page
    list.html         # an index/listing page
    tag.html          # a single tag's listing page
    series.html       # a single series' ordered listing page
  static/             # copied verbatim: css/, images/, fonts/
  public/             # build output (gitignored)
```

The engine defines a small set of **reserved template names** (`base`, `single`, `list`, `tag`, `series`, `404`) and the data each one receives — that naming contract is the API between engine and templates. Every built-in theme provides all of them, so a site needs *zero* templates to build. The lookup runs in two layers: the **selected built-in theme** (`theme.name`) supplies the baseline set, and the site's `layouts/` overrides any of them by name. Document the names and the data each receives, and warn (don't fail) when a file in `layouts/` matches *no* reserved name — that's almost always a typo (`signle.html`) silently overriding nothing.

Two design principles to hold onto:

- **Config-driven, zero hardcoding.** To build *multiple* sites with one binary, nothing about a specific site should live in the tool. The tool reads a config and a directory; everything else is data. The same binary builds any number of separate sites with no code changes.
- **Selectable built-in themes, with per-template site overrides.** The engine embeds one or more built-in themes; `theme.name` in config picks one (default: `"default"`), and a site overrides individual templates *on top of the selected theme*. So the lookup is a three-level cascade resolved per template name: **site `layouts/` → selected built-in theme → (error only if even the theme lacks it, which is an engine bug).** This gives you zero-config sites, theme switching by a single config key, *and* full per-template control. Mechanically it's cheap: selection just parameterizes which embedded subtree feeds the first parse pass (§4). Ship the *selection mechanism* now even if you only author one or two themes at first — the cost is "read the theme dir from config" instead of hardcoding `default`. Each theme bundles its own default stylesheet (and CSS-variable defaults), so every theme looks presentable with zero site CSS; differing *layouts* between themes can come later — the selection plumbing is the part to get in place now.

---

## 4. The parts, and what to build them with

No code — just what each part does and the packages worth using.

### Markdown → HTML
**`goldmark`** is the answer. It's the CommonMark-compliant parser Hugo itself uses, it's fast, and its extension API is clean. Reach for these extensions as you need them:
- GFM (tables, strikethrough, task lists, autolinks) — built in.
- **Syntax highlighting** for code blocks via `goldmark-highlighting` (wraps `chroma`) — built into the engine, driven by config, not template work. See the dedicated subsection below; the light/dark requirement dictates a specific output mode.
- Heading anchors + table-of-contents extensions (e.g. `goldmark-toc`, or roll TOC yourself from the parsed AST).

### Syntax highlighting — built in, themeable, light/dark (decided design)
Decision: **chroma, run by the engine at build time, configured not template-authored.** Shiki's set aside (a JS runtime is too much weight). The non-obvious part is that your *light/dark* requirement forces one specific choice, so get this right early.

**Class-based output, not inline styles — this is the crux.** chroma can emit highlighted code two ways: inline `style="color:…"` on every token, or CSS *classes* (`<span class="k">`, `.s`, `.nf`, …) plus a separate stylesheet. Inline styles bake one theme permanently into the HTML, so **light/dark switching is impossible** without re-rendering. Class-based output is what lets you ship *two* themes and switch between them with CSS alone. So: configure `goldmark-highlighting` with classes on (chroma's `WithClasses(true)`), and have the engine **generate the highlight stylesheet** itself (chroma's HTML formatter can write the CSS for any of its built-in styles).

**How light/dark actually works.** From config you pick two chroma styles — say a light one and a dark one. At build time the engine generates CSS for *both* and scopes them so the page can switch:
- A `prefers-color-scheme` media query gives you automatic OS-based switching with zero JavaScript.
- A `[data-theme="dark"]` / `[data-theme="light"]` selector gives a manual override when the site sets that attribute on `<html>`.
- Emit both, so unset = follow the system, set = force a choice. That's the robust pattern. One implementation wrinkle to expect: chroma writes rules as `.chroma .k { … }`, so to scope a theme under `[data-theme="dark"]` you'll generate each theme's CSS and prefix its selectors (a small string transform), or wrap one in a media query. Budget an hour for getting that scoping right.

Note the *toggle button itself* (and its tiny localStorage JS) is legitimately site UI, not engine work — but because the engine emits CSS keyed off both `prefers-color-scheme` and `data-theme`, system-based dark mode works out of the box with no button at all, and a site that wants a manual toggle just flips `data-theme`.

**Getting it onto the page without template homework.** You don't want users wiring a `<link>` into their `<head>`. Two ways to honor that, since templates live per-site:
- *Auto-injection (most "out of the box").* The engine writes the generated stylesheet to the output (e.g. `assets/syntax.css`) and automatically injects the `<link>` into every page's `<head>` during the write phase. Zero template awareness; the engine fully owns its own asset. This best matches your intent — at the cost of the engine doing a small `<head>` mutation.
- *Provided template variable + scaffolded default.* The engine exposes something like `{{ .GeneratedStyles }}`, and your `new site` scaffold's `base.html` already includes it. The user gets it for free because the starter template has it; a user writing templates from scratch adds one line. Less magic, more idiomatic Go templating.

Either is defensible. Given you explicitly don't want highlighting left to the user, I'd lean **auto-injection** for highlighting specifically, and keep the template-variable mechanism around for anything genuinely optional.

**Config surface** (a `[highlight]` block in `config.toml`):
- `theme_light` / `theme_dark` — chroma style names.
- `line_numbers`, optional per-line highlighting, optional `tab_width`.
- Validate the style names at config load against chroma's known styles, and consider a `yourssg themes` command that lists them — a small DX win that saves a typo-driven debugging session.

**Why this stays simple:** highlighting now happens *inline* during markdown→HTML (a goldmark extension), and CSS generation is a cheap once-per-build step. No second pipeline stage, no Node, no caching needed, still a single static binary. All the complexity that the shiki path introduced disappears.

Also align the engine's syntax CSS with whatever `data-theme` convention the *site's* overall light/dark uses, so the code-block theme and the page theme switch together rather than fighting each other.

### Frontmatter
Two clean options:
- A goldmark frontmatter extension (e.g. **`goldmark-frontmatter`** / `goldmark-meta`) so parsing happens in one pass.
- Or split it yourself: read the file, peel off the `---`-delimited block, decode it with **`gopkg.in/yaml.v3`** or **`BurntSushi/toml`**, then hand the remaining body to goldmark.

Define a frontmatter *schema* early: `title`, `date`, `draft`, `tags`, `slug`, `description`, plus `series` (the series name a post belongs to) and `series_order` (its explicit position within that series). Decode into a typed struct, not a `map[string]any` — you'll thank yourself.

**Derived fields (reading time).** Some page data isn't authored in frontmatter — it's *computed* during the model phase and hung off the page struct for templates. **Reading time** is the canonical example: count the words in the body, divide by an average reading speed (200–225 wpm is the usual range), and round up to whole minutes with a floor of 1. Compute the word count by walking goldmark's parsed text nodes (cleaner than regex on the raw markdown, since it skips syntax and, if you want, code-fence contents). Expose `ReadingTime` and `WordCount` on the page model so a template can render "5 min read." It's computed in Go on the AST, independent of the rest of the pipeline.

### Templating
Start with the **standard library `html/template`**. It's genuinely good: context-aware auto-escaping (security for free), template composition via `{{define}}`/`{{template}}`/`{{block}}`, and a `FuncMap` for custom helpers. For a small SSG it's all you need.
- If you want richer helpers, **`Masterminds/sprig`** adds a big library of template functions.
- If you later want compile-time type safety, **`a-h/templ`** is the modern Go option — but it's a different mental model (templates compile to Go). Don't start there; reach for it only if stringly-typed templates start hurting.

Design your template layer around *layout inheritance*: a `base.html` defines blocks (head, content, footer); `single.html` and `list.html` fill them in. This is how you avoid repeating the page shell.

**The built-in + override mechanism (small, and idiomatic).** Because you now ship built-in templates *and* allow per-site overrides, you need a resolution step — but `html/template` gives it to you almost for free. Parse the **embedded built-ins first** into one template set, then parse the **site's `layouts/` into the same set**. Re-parsing a template of the same name replaces the earlier one, so a site file simply *wins* over the built-in of the same name, and anything the site doesn't provide keeps the built-in. That single "built-ins then site, same set" ordering gives you:
- **Per-file override** — site's `single.html` replaces the built-in `single.html`.
- **Per-block override** — a site can override just a named `{{define "..."}}` block (say, the footer) while inheriting the rest of the built-in shell, since blocks are named templates too.
- **Zero-config builds** — a site with no `layouts/` runs entirely on built-ins.

Load the embedded set via `embed.FS` + `template.ParseFS`, reading from the **selected theme's subtree** (`themes/<theme.name>/`), then layer the site's files on top with another `ParseFS`/`ParseGlob` into the same `*template.Template`. Validate after loading: `theme.name` must name a real embedded theme (error with the list of available ones if not), and every reserved name must resolve to *something* — if one doesn't, that's a bug in that theme, not a user error. Separately, warn on any site file whose name matches no reserved template (the typo case).

If you bundle a default stylesheet with the built-in theme, apply the same fallback idea to that asset: ship the built-in CSS, but let a site's own `static/` file at the same path replace it.

### Config (and config-driven theming)
Use **`spf13/viper`** for configuration. It decodes TOML/YAML into a typed struct, layers in environment variables and defaults, and binds to flags — and it pairs naturally with `spf13/cobra` if you take that route for the CLI, so the config-handling pattern carries over to other CLI projects. (Plain `BurntSushi/toml` or `yaml.v3` decoding is a lighter alternative if you ever want to drop the dependency, but Viper's layering and defaults are convenient here.) Keep every key optional with sensible built-in defaults, so a zero-config site still builds and looks fine.

But config is more than a few scalars here — it also *themes* the site (fonts, colors, syntax theme, default mode). There are two distinct delivery channels, and using each for the right thing is what keeps this clean:

**1. Expose the whole config to templates as data (`.Site`).** Parse the config once and hand it to every template render as something like `.Site`. Now any template — built-in or override — can read `.Site.Title`, `.Site.Author`, `.Site.Nav`, social links, and so on. This is the channel for anything that's *content or markup*, not a CSS value (e.g. building a Google-Fonts `<link>`, rendering nav items). Config-as-data is also your future-proofing for the open-ended "etc.": new config keys are usable in templates with no engine changes.

**2. Drive the built-in theme's CSS with custom properties.** Write the built-in stylesheet against CSS variables — `--font-heading`, `--font-body`, `--font-mono`, `--color-bg`, `--color-text`, `--color-accent`, etc. — and at build time generate a small `:root { … }` block that sets those variables *from config*. The static CSS consumes the variables; the generated block supplies the values. This is the clean way to make fonts/colors configurable without templating or regenerating the whole stylesheet, and it unifies perfectly with light/dark: put light values in `:root` and dark overrides under `[data-theme="dark"]` / `@media (prefers-color-scheme: dark)`. So **site colors, syntax theme, and default mode all switch together through the one `data-theme`/`prefers-color-scheme` convention** you already established for highlighting — that coherence is the payoff.

A representative shape (don't treat as prescriptive — design your own keys):
```
[fonts]
heading = "sans"      # a named preset, a family string, or a web-font name
body    = "serif"
mono    = "mono"      # used by code blocks
source  = "system"    # system | google | self — see the gotcha below

[theme]
name   = "default"    # selects an embedded built-in theme
mode   = "system"     # light | dark | system (default color mode)
accent = "#3b82f6"    # palette overrides; the selected theme supplies defaults

[highlight]
theme_light = "github"
theme_dark  = "github-dark"
```

**The font-loading gotcha (the one people miss).** A `font-family` value does nothing unless the font is actually *available* to the browser. So `fonts.heading = "Inter"` only works if Inter gets loaded. Decide how you support that, ideally all three with `fonts.source`:
- **`system`** — map named presets (`sans`/`serif`/`mono`) to good system-font stacks. Zero network, zero loading, instant. Make this the default; it covers most needs and keeps a static site genuinely static.
- **`google`** — the engine generates the `<link>` (with `preconnect`) from the configured font name and sets the variable. Easy, but it's a third-party request with privacy/perf cost — flag that to users.
- **`self`** — the user drops font files in `static/fonts/` and the built-in theme's `@font-face` references them. Best perf/privacy, more setup.

The font *family string* is a CSS variable (channel 2); the font *loading* (`<link>` or `@font-face`) is channel 1 (generated `<head>` markup or built-in CSS) — that split is exactly why you need both delivery channels.

**Built-in theme selection.** `theme.name` chooses which embedded theme supplies the baseline templates and CSS-variable defaults; `mode`, `accent`, and `[fonts]` then tune whichever theme is selected. Validate `theme.name` against the embedded set and error with the list of available themes on a miss. For v1 the *selection mechanism* is the deliverable even if you only author one or two themes — what matters is that the path is config-driven, not hardcoded.

**Validation & DX.** Validate enumerated values at load: `theme.name` against the embedded themes, `mode` ∈ {light, dark, system}, highlight themes against chroma's known styles, font presets against your named set. A bad value should fail with a helpful message (and a list of valid options), not render a broken site. Mind a naming collision here: you now have *two* kinds of "theme" — built-in site themes and chroma highlight styles. Don't overload one `themes` command for both; e.g. `yourssg themes` lists built-in site themes and `yourssg highlight themes` (or `themes --highlight`) lists chroma styles. Pick a scheme and keep it unambiguous.

### CLI
Your commands will roughly be `build`, `serve`, and `new`.
- The stdlib **`flag`** package plus a small subcommand dispatch (switch on `os.Args[1]`) is enough and has zero dependencies. Very much in the spirit of "small and simple."
- If you want nicer subcommands, flags, and help output, **`alecthomas/kong`** (declarative, struct-tag based, lightweight) or **`spf13/cobra`** (the heavyweight, what Hugo/kubectl use). Cobra is more than you need; kong is a nice middle ground.

### Dev server + live reload
- **`fsnotify/fsnotify`** to watch the content/layouts/static dirs for changes. Add **debouncing** — editors fire multiple events per save, and you don't want a rebuild storm.
- Serve `public/` with `http.FileServer`.
- For live reload, **Server-Sent Events (SSE)** is the simplest path and needs no dependencies: inject a tiny `<script>` (only in dev builds) that opens an `EventSource`; when a rebuild finishes, push an event and the page reloads. Websockets (e.g. `coder/websocket`, `gorilla/websocket`) work too but are overkill for one-way reload signals.

### Small but important utilities
- **`gosimple/slug`** for turning titles into URL-safe slugs.
- **`tdewolff/minify`** for HTML/CSS/JS minification (a later optimization).
- **`golang.org/x/sync/errgroup`** for concurrent page rendering with clean error propagation (see §8).

---

## 5. Suggested build order (this is the "where do I start")

This is the part that matters most when you have "no clue where to start." Build in thin, working vertical slices. After each phase you have something that runs.

**The v1 scope is Phases 0–7** — the live-reload dev server, tags, and series are all in. The phases are ordered by *dependency*, not priority: even though the dev server is a must-have, it can't come first because there has to be something to serve and reload. So it lands at Phase 4, right after the core build works — that's "in v1," just not "first."

- **Phase 0 — Skeleton.** `cmd/yourssg` + a `build` command that loads config and walks `content/`, printing what it finds. No output yet. Goal: the plumbing runs.
- **Phase 1 — Markdown + frontmatter.** Parse one file: split frontmatter, decode it, convert the body to HTML with goldmark. Print the HTML. Goal: content parsing works.
- **Phase 2 — Templates + write.** Load the site's `base.html`/`single.html`, execute against a page, write a real HTML file to `public/`. Now you can open it in a browser. Pass the parsed config in as `.Site` from the start, so templates can read site-level values (title, author, nav) immediately. **This is your first real site.**
- **Phase 3 — Whole tree + assets.** Render *all* content files to the right output paths, copy `static/` verbatim, handle the base URL and pretty URLs (`/posts/hello/` via `hello/index.html`). Goal: a full, browsable static site.
- **Phase 4 — `serve` with live reload (v1 must-have).** fsnotify (debounced) + rebuild + SSE. This single feature transforms the development experience; it's a v1 must-have, so build it as soon as the build pipeline is solid rather than letting it drift later.
- **Phase 5 — Content model: tags + series + reading time.** Build the full collection model *before* rendering (see §8 — series needs every page discovered first so prev/next links resolve). Generate a listing page per tag (`tag.html`) and per series (`series.html`, ordered by `series_order`), and expose each page's tags and series neighbours to templates. Compute **reading time** here too (word count off the AST, ÷ ~200 wpm, ceil to whole minutes). Drafts (skip `draft: true` unless a flag is set) and future-dated handling fit here. **Pagination is explicitly deferred** — a long list page is fine for now.
- **Phase 6 — Highlighting + config-driven theming + content polish + web essentials.** Wire up built-in syntax highlighting (chroma, class-based, configured themes): emit classes, generate the light/dark stylesheet, auto-inject it (§4). Generate the **theme-variable `:root` block from config** (fonts, colors, default mode) and any font-loading `<link>`/`@font-face`, so the built-in CSS becomes configurable (§4 Config). Then heading anchors, TOC, RSS/Atom, `sitemap.xml`, `robots.txt`, SEO/Open Graph meta tags, a `404.html`. All single-binary, single-stage — no second pipeline, no Node.
- **Phase 7 — CLI polish.** A `new` command — `new post` (scaffold a content file with prefilled frontmatter) and `new site` (scaffold `config` + a sample post; templates are now optional since the built-ins cover a fresh site). An `eject` command that copies the **selected** theme's templates into `layouts/` so a user can customize from a known-good base instead of a blank file. List commands that respect the naming split: `themes` for built-in site themes, `highlight themes` (or `themes --highlight`) for chroma styles. Nicer help/flags (kong/cobra if you went that way).
- **Phase 8 — Distribution.** GoReleaser + a GitHub Action wrapper (see §6).
- **Phase 9 — Deploy targets.** Workflows for Pages / Cloudflare / Vercel (see §7).
- **Phase 10 — Optimizations.** Concurrent rendering, minification, asset fingerprinting, incremental builds. *Only after the rest works and you've measured.*

Resist the urge to build Phase 10 things early. The first time you "optimize" before there's anything to optimize, you'll regret it.

---

## 6. Making the CLI usable in GitHub Actions (across many sites)

You have four distribution mechanisms, roughly in increasing order of polish:

1. **`go install`.** In a workflow: `actions/setup-go`, then `go install github.com/<you>/yourssg@latest`, then run it. Zero release infrastructure. Downside: every CI run compiles the tool (slow-ish), and you depend on a Go toolchain in CI.

2. **Prebuilt release binaries via GoReleaser.** **`goreleaser/goreleaser`** cross-compiles for linux/macOS/windows × amd64/arm64, builds archives + checksums, and attaches them to a GitHub Release on every tag push. Embed the version via `-ldflags`. CI then downloads the right binary instead of compiling. This is the foundation for everything below.

3. **A composite GitHub Action (the key one for "multiple sites").** Add an `action.yml` at the repo root that downloads your prebuilt binary (from the GoReleaser release matching the action version) and runs the build. Then *any* of your site repos uses it with one line:
   ```yaml
   - uses: <you>/yourssg@v1
     with:
       config: ./config.toml
   ```
   This is the cleanest "build it once, use it everywhere" story. Tag the action with SemVer and a moving `v1` tag so consumers can pin a major version.

4. **A Docker-based action.** Publish an image to **GHCR** (`ghcr.io/<you>/yourssg`) and make the action `runs.using: docker`. Fully hermetic, no toolchain assumptions, but slower to pull and a heavier thing to maintain. Worth it only if you hit environment-drift pain.

**Recommended path:** GoReleaser (for binaries) + a composite `action.yml` wrapping it. That gives you SemVer'd releases *and* a one-line `uses:` for every site, without Docker overhead. With highlighting handled in-process by chroma, your build stays a single static binary and needs no Node or other runtime in CI — so this composite-action path is unambiguously the right one; the Docker-based option (4) is now only an "if you ever hit environment drift" fallback, not something the toolchain forces on you.

A couple of details that bite people:
- **Version stamping**: inject version/commit/date via ldflags so `yourssg version` is meaningful and your action can verify it downloaded the right build.
- **Templates: built-ins are in the binary, overrides come from the site repo.** The default theme is `go:embed`'d, so any site builds in CI with no template files at all. If a site provides overrides in `layouts/`, CI already has them — it checks out the site repo before running your tool. Either way the binary stays self-contained; just run the action from the site's root so relative paths (`content/`, `layouts/`, `static/`) resolve.

---

## 7. Deployment (Pages, Cloudflare, Vercel, and the general case)

The shape is always the same: **CI builds `public/`, then a host serves it.** What differs is the last step.

**GitHub Pages.** Two routes:
- The official flow: `actions/upload-pages-artifact` to package `public/`, then `actions/deploy-pages` to publish. This is the modern, recommended path.
- The classic flow: `peaceiris/actions-gh-pages` to push `public/` to a `gh-pages` branch.
- **The gotcha:** project pages live under `https://<you>.github.io/<repo>/`, so your **base URL has a path prefix**. Every internal link and asset reference must respect it, or everything 404s. Make base URL a config value and set it per environment. (User/org pages at the root domain don't have this problem.)

**Cloudflare Pages.**
- Connect the repo in the dashboard and set build command + output dir (`public`), letting Cloudflare build on push; or
- Build in your own Actions workflow and deploy with **`wrangler pages deploy public`**, authenticating via `CLOUDFLARE_API_TOKEN` + `CLOUDFLARE_ACCOUNT_ID` secrets. The wrangler route gives you full control over the build environment.

**Vercel.** Set the framework preset to "Other", build command to your build, output dir to `public`. Either let Vercel auto-deploy on push (repo connected) or run the **Vercel CLI** / a community action in your own workflow with a `VERCEL_TOKEN`. Vercel's preview-deploy-per-PR is a nice bonus for content review.

**Netlify / generic.** Same idea: build command + publish dir. And because the output is *just files*, you can drop it on anything — Cloudflare R2 + a custom domain, S3 + CloudFront, a plain nginx box, whatever.

Cross-cutting deployment concerns to bake into the generator, not the host:
- **Base URL per environment** (prod vs. preview vs. the GitHub Pages path prefix).
- **Trailing-slash / pretty-URL strategy** — pick one and be consistent (`/posts/hello/` with `index.html` is the common choice).
- **`404.html`** — most hosts will serve it automatically for unknown paths.
- **`sitemap.xml`, `robots.txt`, RSS** — generate them at build time.
- **Canonical URLs + Open Graph tags** — needs the absolute base URL, another reason it's config.

---

## 8. Things you might have missed

A grab-bag of things that separate a toy from something you'll actually want to use:

- **Content model design.** Decide up front how frontmatter fields behave: drafts, future-dated posts, tags/categories (taxonomies), per-page slug overrides, descriptions. The model is the spine of the whole tool.
- **Pretty/permalink URLs.** `content/posts/hello.md` → `/posts/hello/`. Decide the mapping rules and make them predictable.
- **List pages & pagination.** An index of all posts, per-tag listings, and pagination once a list gets long.
- **Taxonomies — tags vs. series (the v1 collections).** Both are "group pages by a frontmatter field," so they share machinery, but they differ in two ways. **Tags** are *unordered* and *many-per-page*; each tag term becomes a listing page (`tag.html`), and a page links out to all its tags. **Series** are *ordered* and usually *one-per-page*: each series becomes an ordered listing page (`series.html`, sorted by `series_order`), and — the part tags don't have — each member page needs **prev/next** links and "Part N of M" context. The key architectural consequence: you must build the *complete* content model (all pages discovered, all collections assembled and sorted) **before rendering any page**, because page N's "next" link can't resolve until page N+1 exists in the model. That's a hard ordering barrier between the Model and Render phases — and, incidentally, the main thing that constrains how far you can parallelise rendering (see the concurrency note below: render is parallel, but only *after* the model is fully built).
- **Pagination (deferred for v1).** Tag and series listings will eventually get long; pagination is "chunk a sorted list into N-per-page with `page/2/` URLs." It's out of scope for v1 — just keep the listing-page rendering structured so adding a paginator later doesn't mean a rewrite.
- **RSS/Atom, sitemap, robots, SEO meta.** Small to add, big for a real site. Open Graph/Twitter cards make shared links look right.
- **Heading anchors & TOC.** Quality-of-life for technical/long posts.
- **Syntax highlighting — built in and themeable (§4).** chroma, class-based output (the only mode that allows light/dark), engine-generated stylesheet for a configured light + dark theme, auto-injected so users never touch templates. Single binary, single stage, no caching needed. The one fiddly bit is scoping each theme's CSS under `prefers-color-scheme` and/or `[data-theme]`.
- **Reading time (and word count).** Computed, not authored — derive it in the model phase from an AST word count ÷ ~200 wpm, ceil to whole minutes, floor of 1. Hang it on the page struct for templates. Cheap, and a nice signal on technical posts.
- **Config-driven theming (§4).** Two channels, used for the right things: expose the whole config as `.Site` data for markup, and drive the built-in CSS with generated `:root` custom properties for fonts/colors/mode. Everything (site palette, syntax theme, default mode) switches together through the one `data-theme`/`prefers-color-scheme` convention. The gotcha worth a sticky note: a `font-family` value does nothing unless the font actually *loads* — support system stacks by default, with `google`/`self` as opt-in via a `fonts.source` key.
- **Dev server with live reload.** The single biggest DX multiplier, and a v1 commitment for you — Phase 4, don't let it slip.
- **Good build errors.** When a template fails or frontmatter is malformed, report the *file and line*. Cryptic errors make a tool miserable to use; this is where a personal tool can beat the big ones.
- **Testing strategy.** **Golden-file tests** fit SSGs perfectly: feed a known input tree, compare the produced `public/` against a committed expected tree. Table-driven tests for markdown/frontmatter edge cases. This is also where your instinct to understand bugs conceptually pays off — golden tests pin behavior precisely.
- **Concurrency.** Page rendering is embarrassingly parallel — independent pages, no shared mutable state if you design it right. A worker pool or `errgroup` over the page list parallelizes rendering cleanly. *Caveat:* much of the work is I/O (reading files, writing output), so measure before assuming CPU parallelism helps. The interesting concurrency design question is the *write* phase and any shared caches (e.g. a parsed-template cache behind an `RWMutex`).
- **Asset pipeline (later).** Minification (tdewolff/minify) and **fingerprinting** (hash file contents, rename `style.css` → `style.abc123.css`, rewrite references) for cache-busting. Defer until the basics are solid.
- **Incremental builds (much later).** Cache by content hash; only re-render what changed. Real complexity for real payoff on big sites — and a genuinely interesting systems problem if you want one. Not a v1 concern.
- **Template resolution & typo warnings.** With built-ins as a fallback, a missing reserved template is no longer a user error — the built-in fills in. The failure mode flips: a site file that matches *no* reserved name (a typo like `signle.html`) silently overrides nothing and the user wonders why their change did nothing. Warn on unrecognized template names in `layouts/`. And after loading, assert every reserved name resolves to something — if not, that's a bug in your embedded theme, surfaced early rather than as a nil panic mid-render.
- **Shortcodes / render hooks (advanced).** Hugo's shortcodes let content embed reusable snippets (e.g. a callout, a figure). Powerful but a rabbit hole — only if you find yourself wanting it.

---

## 9. Deep dive: Series (and the model-before-render barrier)

Series is the feature that quietly dictates your build architecture, so it's worth understanding completely before you write the model code.

### The data you're building toward
Two structures. A **Series** has a name, a slug (for its URL), a display title, and an *ordered* list of its member pages. A **Page**, in addition to its authored frontmatter (`series` name + `series_order`), gains a set of *derived* fields once the model is assembled: its position in the series (N), the series total (M), and pointers (or nil) to the previous and next pages in reading order. Templates consume those derived fields; your job in the model phase is to populate them correctly.

### The three passes
Building this cleanly is three sequential passes over your pages:

1. **Discover & parse** — read every content file, parse frontmatter and body. You already do this. At the end you have a flat list of all pages with their `series`/`series_order` known but no adjacency yet.
2. **Group** — bucket pages by their `series` field into a map of series-name → pages. This is the exact same "group by a frontmatter field" operation tags use; factor it into one helper and let tags and series share it. (Tags stop here — a tag bucket is just an unordered set.)
3. **Sort & link** — for each series bucket, sort the members by `series_order`, then walk the sorted slice assigning each page its position, the total, and prev/next pointers. This adjacency-assignment walk is the only thing series adds over tags.

### Why this forces a hard barrier
A single post's template wants to render "Next: *Part 3 — Parsing the header*" with a working link. To produce that, the post needs its *successor's* title and URL — which don't exist until the successor has been discovered, parsed, and placed in the sorted series. Therefore **every page must be fully modelled before any page is rendered.** This is the model-before-render barrier, and it's not optional once you have series.

The upside is large: once the model is complete it's effectively *read-only*, which is exactly what makes the render phase safe to parallelise. Each page renders independently against a frozen, shared model — no locks, no ordering between renders. So the barrier you're "forced" into is also what unlocks clean concurrency later (§8). Structure your `build` orchestrator as two distinct stages with the barrier between them: `buildModel()` returns a finished, immutable site model; `render(model)` fans out.

### Ordering: explicit, not by date
Use an explicit integer `series_order`, not publish date. You'll revise part 2 after part 4 is out; you'll backfill; you'll reorder. Date-based ordering makes all of that fragile. Leave **gaps** (10, 20, 30…) so you can insert a part later without renumbering everything. Two determinism rules to bake in:
- **Tie-break.** If two pages share an order (a mistake, but it happens), fall back to date, then filename, so the build is reproducible rather than dependent on filesystem walk order. Emit a warning when you detect a duplicate order — it's almost always an authoring error.
- **Stable map iteration.** Go randomizes map iteration. When you iterate the series (or tag) map to render index pages, sort the keys first. Otherwise your output reorders between builds, which wrecks golden tests and produces noisy `gh-pages` diffs.

### The draft/future-dated gotcha (this is the subtle one)
Drafts and future-dated posts get filtered out of production builds. If you filter *after* assigning adjacency, you get a dangling "Next →" link pointing at a post that was never published — a 404 in your reader's face. So the filter must happen **before** the sort-and-link pass: remove non-visible pages from the series first, *then* compute positions and prev/next over the visible set only. That also keeps "Part N of M" contiguous — readers see Parts 1, 2, 3, not 1, 2, 4 with a hole. (In `serve`/dev mode you'll typically include drafts, so adjacency naturally differs between dev and prod — that's correct, not a bug.)

### Edge cases to handle
- **Single-member series** → no prev and no next. Don't let templates nil-panic; expose `HasPrev`/`HasNext` booleans (or guarantee nil-safe access) so a one-part "series" renders fine.
- **Missing `series_order`** on a page that declares a `series` → either hard-error at model time (recommended; cheap and unambiguous) or document a date fallback. Pick one and be loud about it.
- **One series per post for v1.** Keep `series` a single string. Multi-series membership (a list, a page holding several positions) is a real feature but a needless complication now — defer it.
- **The series index page** lives at `/series/<slug>/` and renders with `series.html` over the ordered members. Member posts keep their normal `/posts/...` URLs; the index just links to them. Don't republish post bodies under the series path — that's duplicate content.

### A concrete walk-through
Say you have a three-part series — call it `getting-started` — with `series_order: 10, 20, 30`. Pass 1 parses all three (plus every unrelated page). Pass 2 buckets the three under `getting-started`. Pass 3 sorts them 10→20→30, then walks: part 10 gets `{pos:1, total:3, prev:nil, next:→20}`, part 20 gets `{pos:2, total:3, prev:→10, next:→30}`, part 30 gets `{pos:3, total:3, prev:→20, next:nil}`. Now — and only now — rendering can begin, and each page's template can render its series nav. If part 20 were a draft in a prod build, you'd filter it first, and the visible set becomes 10→30 with totals of 2, so the numbering stays honest.

---

## 10. Deep dive: Live reload (the dev server)

`serve` is where a few well-known traps live. The mechanism itself is small; the traps are what bite. Here's the whole thing.

### The shape of `serve`
One command wires together four moving parts: an **HTTP server** over your output dir, a **file watcher** on your sources, a **rebuild** trigger, and a **notifier** that tells connected browsers to reload. The loop is: do a full build, serve it, watch sources, and on change → rebuild → broadcast "reload" → browsers refresh.

### Serving the files
Point `http.FileServer` at the output dir. It already does the pretty-URL thing — a request for `/posts/hello/` serves `/posts/hello/index.html`. Two dev-only behaviours to add via middleware: inject the live-reload script (below), and send **`Cache-Control: no-store`** on everything. That cache header is a classic miss — without it, the browser serves stale CSS/JS from memory even after the reload fires, and you'll swear your changes "aren't taking" when they actually built fine.

### Watching with fsnotify — two traps
Use `fsnotify`, and watch your *source* directories: `content/`, `layouts/`, `static/`, and the config file.

- **Trap 1 — never watch the output dir.** Your build writes into `public/`. If `public/` is watched, every build triggers the watcher, which triggers a build, forever. Exclude it explicitly. (Belt and suspenders: also ignore events whose path is inside the output dir.)
- **Trap 2 — fsnotify isn't recursive.** It watches a directory, not its subtree. You must walk the source tree at startup and `Add` every subdirectory, and when a *new* directory is created you must `Add` that too (watch for create events on dirs). Miss this and new content folders silently won't trigger rebuilds.

### Debouncing the event storm
A single editor save does not produce a single event. You'll see CREATE + WRITE + CHMOD, atomic-save temp-file + rename dances, vim swap/backup files, etc. — often several events in a few milliseconds. If you rebuild per event you'll rebuild five times per save. **Debounce:** on each relevant event, (re)start a short timer (≈100–150ms); only when it fires with no further events do you rebuild. The reset-timer pattern (`time.AfterFunc` reset, or a select loop draining an event channel with a timeout) collapses the burst into one rebuild. While you're there, filter out noise — editor swap files, dotfiles, anything under the output dir — before the timer even starts.

### Rebuild strategy: full, for now
On a triggered change, just rebuild the whole site. For a personal site this is milliseconds and you keep the dev path identical to the prod path (fewer "works in build, breaks in serve" surprises). *Incremental* rebuilds — only re-render what changed — are genuinely hard because the dependency graph is wide: a template edit invalidates everything; a single post edit invalidates that post *plus* every list/tag/series page that references it *plus* its series neighbours' nav. That's a Phase 10 problem, and only if full rebuilds ever feel slow. Don't pay for it now.

### Notifying the browser: SSE, not websockets
The browser only needs a one-way "reload now" signal, so **Server-Sent Events** is the right tool — plain HTTP, no dependencies, and `EventSource` reconnects automatically if the dev server restarts. Websockets are bidirectional and bring a library; that's overkill for a reload ping.

The server side is a small **broker/hub**: an endpoint (say `/__livereload`) sets `Content-Type: text/event-stream`, registers a per-connection channel into a client set, and streams whatever lands on that channel. On rebuild you broadcast a "reload" message to every registered channel. Guard the client set with a mutex, or use the idiomatic Go pub-sub shape — a single goroutine owning the set with `register`/`unregister`/`broadcast` channels and a `select` loop. Clean up on disconnect by watching `r.Context().Done()` and unregistering, or you'll leak a channel per page refresh. Send a heartbeat comment line (`: ping`) every 15–30s so idle connections aren't dropped by the browser or any proxy.

### Injecting the client script — dev only, in flight
The browser needs a few lines of JS that open `new EventSource('/__livereload')` and call `location.reload()` on a message. The clean way to deliver it: have your dev-server middleware splice the `<script>` into `text/html` responses just before `</body>` **as they're served** — not into the built files, and not into your templates. This keeps your `build` output pristine (no live-reload cruft ships to production) and keeps templates ignorant of dev concerns. It mirrors the auto-injection idea from highlighting, but strictly dev-only.

### The ordering rule that makes it feel instant (and correct)
Sequence matters: watcher fires → debounce → **rebuild to completion** → *then* broadcast reload. If you broadcast before the files are written, the browser reloads stale output. And if the rebuild *fails*, do **not** broadcast a reload — keep serving the last good build and surface the error (log it at minimum; a nice touch is an error overlay pushed over SSE). Reloading onto a broken build just flashes a white page and hides the actual compile/template error.

### Putting it together
Three goroutines and a server: the HTTP server (FileServer + the `/__livereload` SSE endpoint + the HTML-injection/no-cache middleware); the watcher goroutine (fsnotify → debounce → rebuild → broker.broadcast); and the broker goroutine (owns the SSE client set). Wire Ctrl-C to a context cancel that stops the watcher, closes client channels, and shuts the server down gracefully. End-to-end the data flows: filesystem event → debounce timer → `rebuild()` → `broker.broadcast("reload")` → each client channel → browser `EventSource` → `location.reload()`.

### The gotcha checklist
- Don't watch the output dir (infinite loop).
- Walk + `Add` subdirectories; add newly created dirs too (fsnotify isn't recursive).
- Debounce, and ignore editor/swap/dotfiles.
- Rebuild fully to completion *before* broadcasting.
- On build failure: don't reload; keep last good output; show the error.
- Inject the reload script dev-only and in-flight, never into build output.
- `Cache-Control: no-store` in dev, or stale assets will fool you.
- SSE heartbeat + clean unregister on disconnect (no leaked channels).
- Handle "port already in use" (flag to set the port, or auto-increment and announce it).

---

## A note on scope

The fastest way to never ship this is to chase Hugo feature-for-feature. The guiding constraint — *HTML templates, CSS, and markdown, nothing more* — is the right one. v1 is Phases 0–7, but you don't have to reach Phase 7 before seeing anything real: get Phases 0–4 building and live-reloading a real site first, deploy that early cut, then fill in tags, series, reading time, and highlighting as you go. A small SSG that builds a handful of sites perfectly beats a large one that tries to build everyone's.

