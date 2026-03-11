# repo-tokens

Zero-config token counting and context-window badges for any codebase.

[![tokens](.github/badges/tokens.svg)](https://github.com/ehmo/repo-tokens)

## Install

```bash
go install github.com/ehmo/repo-tokens/cmd/repo-tokens@latest
```

Or download a binary from [Releases](https://github.com/ehmo/repo-tokens/releases).

## Quick Start

```bash
repo-tokens                     # Count tokens in current directory
repo-tokens ~/my-project         # Count a specific path
repo-tokens init                 # Set up GitHub Action + badge
```

## Usage

```bash
repo-tokens [flags] [path]              # Count tokens
repo-tokens init [path]                  # Set up workflow + badge for a repo

# Flags
--badge badge.svg                        # Generate SVG badge
--top 10                                 # Show top 10 files by token count
--json                                   # JSON output
--detect                                 # Show detected projects only
--include-tests                          # Include test files
--context-window 128000                  # Override context window (default: 200k)
--update-readme README.md                # Update token markers in README
--marker token-count                     # Custom marker name
```

## How it works

1. Detects project types from markers (`go.mod`, `package.json`, `*.xcodeproj`, etc.)
2. Also scans by file extension to catch languages without markers (e.g. C++ in a repo with a Python build system)
3. Falls back to extension-only detection if no markers found
4. Counts tokens with [tiktoken](https://github.com/openai/tiktoken) (cl100k_base encoding)
5. Generates a color-coded badge based on context window fill

### Monorepo support

Finds and counts separate projects automatically:

```
$ repo-tokens ~/my-monorepo

  my-monorepo — 2 projects (context: 200k)

  PROJECT                  TYPE    FILES   TOKENS   CTX
  ──────────────────────────────────────────────────────────────
  apps/ios                swift      133     332k   166%  [████████████████████]
  apps/web                  web       57     196k    98%  [███████████████████░]
  ──────────────────────────────────────────────────────────────
  Total                              190     528k   264%
```

### Find bloat

```
$ repo-tokens --top 5

  Top 5 files by token count:

     29.1k  src/storage/BackupManager.swift
     14.5k  src/features/ViewModel.swift
     13.6k  src/storage/Storage.swift
     12.5k  src/sharing/CloudKit.swift
     11.3k  src/sharing/Serializer.swift
```

## Supported languages

| Type | Markers | Extensions |
|------|---------|------------|
| Go | `go.mod` | `.go` |
| Rust | `Cargo.toml` | `.rs` |
| C/C++ | `CMakeLists.txt` `meson.build` `configure.ac` `Kconfig` | `.c .h .cpp .hpp .cc .cxx .hh` |
| Zig | `build.zig` | `.zig` |
| Nim | `*.nimble` | `.nim .nims` |
| V | `v.mod` | `.v` |
| Web/JS | `package.json` | `.js .ts .jsx .tsx .mjs .mts .html .css .scss .sass .less .vue .svelte .njk .liquid .astro .mdx` |
| Swift | `Package.swift` `*.xcodeproj` `*.xcworkspace` | `.swift` |
| Kotlin | `build.gradle.kts` | `.kt .kts` |
| Dart | `pubspec.yaml` | `.dart` |
| Java | `build.gradle` `pom.xml` | `.java` |
| Scala | `build.sbt` | `.scala .sc` |
| Clojure | `project.clj` `deps.edn` | `.clj .cljs .cljc .edn` |
| Python | `pyproject.toml` `setup.py` `setup.cfg` `requirements.txt` | `.py .pyi` |
| Ruby | `Gemfile` | `.rb .erb .rake` |
| PHP | `composer.json` | `.php` |
| Perl | `Makefile.PL` `cpanfile` `dist.ini` | `.pl .pm .t` |
| Lua | `*.rockspec` `.luacheckrc` | `.lua` |
| C# | `*.csproj` `*.sln` | `.cs` |
| F# | `*.fsproj` | `.fs .fsi .fsx` |
| Elixir | `mix.exs` | `.ex .exs .heex .leex` |
| Erlang | `rebar.config` | `.erl .hrl` |
| Haskell | `*.cabal` `stack.yaml` | `.hs .lhs` |
| OCaml | `dune-project` `*.opam` | `.ml .mli` |
| Gleam | `gleam.toml` | `.gleam` |
| Julia | `Project.toml` | `.jl` |
| R | `DESCRIPTION` `.Rproj` | `.R .r .Rmd` |
| Terraform | `*.tf` | `.tf .tfvars` |

**Extension-only** (detected without markers): Shell, SQL, GraphQL, Protobuf, YAML, TOML, Markdown, LaTeX

## Exclusions

Skips by default:
- Test dirs: `tests/`, `*Tests/`, `*UITests/`, `__tests__/`, `spec/`
- Test files: `*_test.go`, `*.test.ts`, `*.spec.js`, etc.
- Dependencies: `node_modules/`, `vendor/`, `Pods/`, `.build/`
- Build artifacts: `dist/`, `build/`, `DerivedData/`, `target/`
- Generated files: `*.min.js`, `*.min.css`, `output.css`, `*.bundle.js`
- Lock files: `package-lock.json`, `yarn.lock`, `go.sum`, `Cargo.lock`, etc.
- Binary files (null byte check)
- Hidden dirs: `.git/`, `.svn/`, `.idea/`, etc.

Use `--include-tests` to count test files.

## Badge

Color reflects context window fill:

| Color | Range | Meaning |
|-------|-------|---------|
| Green | < 30% | Agent can hold full codebase |
| Yellow-green | 30-50% | Comfortable fit |
| Yellow | 50-70% | Getting tight |
| Red | > 70% | Agent will struggle |

Default context window: 200k (Claude). Override with `--context-window`.

## GitHub Action

### Quick setup

```bash
repo-tokens init
```

This creates the workflow file, adds README markers, and generates a config for monorepos.

### Manual setup

Add to your README:

```html
<!-- token-count --><a href="https://github.com/ehmo/repo-tokens">11.7k tokens · 6% of context window</a><!-- /token-count -->
```

Create `.github/workflows/repo-tokens.yml`:

```yaml
name: Update token count

on:
  push:
    branches: [main]

permissions:
  contents: write

jobs:
  tokens:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: ehmo/repo-tokens@v1
        id: tokens

      - name: Commit if changed
        run: |
          git add -A .github/badges/ README.md
          git diff --cached --quiet && exit 0
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git commit -m "docs: update token count to ${{ steps.tokens.outputs.badge }}"
          git push
```

### Inputs

| Input | Default | Description |
|-------|---------|-------------|
| `path` | `.` | Path to scan |
| `context-window` | `200000` | Context window size |
| `encoding` | `cl100k_base` | Tiktoken encoding |
| `badge-path` | `.github/badges/tokens.svg` | SVG badge output path |
| `readme` | `README.md` | README for marker replacement |
| `marker` | `token-count` | HTML comment marker name |
| `include-tests` | `false` | Include test files |

### Outputs

| Output | Description |
|--------|-------------|
| `tokens` | Total token count |
| `percentage` | Context window percentage |
| `badge` | Formatted badge text |
| `json` | Full JSON with per-project breakdown |

## Config file

Create `.repo-tokens.yml` for custom settings:

```yaml
context-window: 200000

projects:
  - name: Web App
    path: apps/web
    type: web
    exclude: [storybook]

  - name: iOS App
    path: apps/ios
    type: swift
```

## Why

Small codebases were always a good thing. With coding agents, it's now a measurable advantage.

### The context window problem

Every coding agent, whether Claude Code, Cursor, Windsurf, or Copilot, works within a context window. That window is the agent's entire working memory: your code, the conversation, its reasoning, and its output all compete for the same space. The bigger your codebase relative to that window, the worse the agent performs.

The research backs this up. LLM performance degrades well before hitting the technical context limit:

- "Lost in the Middle" (Liu et al., 2023) found that models perform best on information at the beginning and end of context, with accuracy dropping for content in the middle, even for models designed for long contexts. ([arxiv.org/abs/2307.03172](https://arxiv.org/abs/2307.03172))
- RULER (Hsieh et al., 2024) tested 17 long-context models and found only half maintained satisfactory performance at their claimed 32K context. Simple needle-in-a-haystack tests mask the real degradation on complex tasks. ([arxiv.org/abs/2404.06654](https://arxiv.org/abs/2404.06654))
- Levy et al., 2024 showed that adding irrelevant text to inputs causes reasoning performance drops at lengths far shorter than the model's maximum. ([arxiv.org/abs/2402.14848](https://arxiv.org/abs/2402.14848))
- NeedleBench (2024) found that even reasoning models like DeepSeek-R1 and OpenAI o3 struggle with retrieval and reasoning in information-dense scenarios. ([arxiv.org/abs/2407.11963](https://arxiv.org/abs/2407.11963))

The practical takeaway: stay within 50-70% of a model's context window. Beyond that, you're gambling with accuracy.

### Model context windows

Not all context is equal. Providers signal this through tiered pricing, charging 1.5-2x more beyond certain thresholds because performance degrades and compute costs spike.

| Model | Provider | Context Window | Effective Limit | Notes |
|-------|----------|---------------|-----------------|-------|
| Claude Opus 4.6 | Anthropic | 200K | ~200K | 1M beta available; long-context surcharge beyond 200K |
| Claude Sonnet 4.6 | Anthropic | 200K | ~200K | 1M beta available |
| GPT-5.4 | OpenAI | ~500K+ | ~272K | [Price jumps 2x input / 1.5x output after ~272K; degraded performance beyond](https://x.com/olegkhomenko_/status/2031281737867669871) |
| GPT-4.1 | OpenAI | 1M | ~200K | 1M technically available but quality drops on complex tasks |
| o3 | OpenAI | 200K | ~200K | Reasoning model |
| Gemini 2.5 Pro | Google | 1M | ~500K | Long context available but "lost in the middle" applies |
| Gemini 2.5 Flash | Google | 1M | ~500K | |
| Grok 4 | xAI | 256K | ~128K | 2x price beyond 128K threshold |
| Grok 4.1 Fast | xAI | 2M | ~128K | 2x price beyond 128K; massive context ≠ reliable context |
| Kimi K2 | Moonshot AI | 128K | ~100K | 1T total params, 32B activated (MoE) |
| MiniMax-Text-01 | MiniMax | 1M-4M | ~500K | Hybrid Lightning + Softmax attention architecture |

"Effective limit" = where pricing jumps (a signal you're in degraded territory) or research shows quality loss on multi-step reasoning.

### What this means for coding agents

SWE-bench ([arxiv.org/abs/2310.06770](https://arxiv.org/abs/2310.06770)) showed that coding agents need to "understand and coordinate changes across multiple functions, classes, and even files simultaneously" and that context management is the bottleneck. SWE-agent ([arxiv.org/abs/2405.15793](https://arxiv.org/abs/2405.15793)) found that agents with smaller, focused windows consistently outperform those that dump more code into context.

In practice:
- < 30% of context: the agent holds your full codebase with room to reason. Sweet spot.
- 30-70%: code fits but the agent has less room for reasoning, conversation history, and tool output. Quality slips on complex multi-file changes.
- \> 70%: the agent can't hold your full codebase. It selectively loads files, loses track of dependencies, and makes more errors. Every extra token of code means fewer tokens for thinking.

A 50k-token codebase in a 200K window leaves 150K tokens for the agent to think, plan, and iterate. A 180k-token codebase leaves almost nothing. The code might technically "fit" but the agent is lobotomized.

### How I got here

I built [mdgrok.com](https://mdgrok.com), a search engine that indexes `AGENTS.md` and `CLAUDE.md` files from thousands of open-source repos. These are the config files that tell AI coding agents how to navigate a project: what conventions to follow, what to skip, how things are structured.

After looking at 50,000+ of these files across 4,500+ repositories, a pattern jumped out: the projects where AI agents work best aren't just the ones with good `AGENTS.md` files. They're the ones where the agent can actually hold the codebase in its head. A perfect config file doesn't help if the codebase is so large that the agent loses track of what it read three files ago.

That's what led to repo-tokens. I wanted a single number, and a badge, that tells you whether your codebase fits in an agent's working memory. Not the theoretical context window. The *useful* one, where the agent can still reason about your code.

## License

MIT

## Author

Built by [Rasty Turek](https://turek.co).
