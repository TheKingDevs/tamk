# Doc-Audit Skill

Cross-verify documentation against actual code in TAMK. Detects discrepancies where docs describe things that don't exist in code, or code has things docs don't mention.

## When to use

- After major code changes (new features, refactors, type renames)
- Before releases or version bumps
- When user says "T2T", "audit docs", "verify docs", or "doc-code check"

## Workflow

### Phase 1: Inventory (read-only)

Read these files in parallel:

```
AGENTS.md
documentation/*.md (all 21)
internal/config/config.go
internal/domain/entity/*.go
internal/domain/repository/*.go
internal/domain/valueobject/*.go
internal/usecase/*.go
internal/repository/filesystem/*.go
internal/delivery/cli/*.go
pkg/errors/errors.go
pkg/logger/logger.go
pkg/watcher/watcher.go
pkg/qrcode/qrcode.go
templates/**/*.tmpl
Makefile
```

### Phase 2: Cross-verify (15 checks)

For each check, compare what docs claim vs what code actually defines:

| # | Check | Docs Source | Code Source |
|:--|:--|:--|:--|
| 1 | **File structure** | AGENTS.md §Project Structure | `find . -type f` |
| 2 | **CLI commands** | AGENTS.md §CLI Reference | `internal/delivery/cli/root.go` (AddCommand calls) |
| 3 | **Entity types** | AGENTS.md §Domain Layer table | `internal/domain/entity/*.go` (type declarations) |
| 4 | **Use case methods** | AGENTS.md §Use Case Layer table | `internal/usecase/*.go` (func signatures) |
| 5 | **Repository interfaces** | AGENTS.md §Repository Layer table | `internal/domain/repository/*.go` (interface methods) |
| 6 | **Repository implementations** | AGENTS.md §Repository Layer table | `internal/repository/**/*.go` |
| 7 | **Template files** | AGENTS.md §Template System table | `templates/**/*.tmpl` (glob) |
| 8 | **Error sentinels** | AGENTS.md §Errors table | `pkg/errors/errors.go` (var block) |
| 9 | **Watcher extensions** | AGENTS.md §File Watcher | `pkg/watcher/watcher.go` (WatchExtensions map) |
| 10 | **Watcher ignore dirs** | AGENTS.md §File Watcher | `pkg/watcher/watcher.go` (IgnoreDirs map) |
| 11 | **Logger functions** | AGENTS.md §Color Utilities table | `pkg/logger/logger.go` (exported funcs) |
| 12 | **QR code functions** | AGENTS.md §QR Code table | `pkg/qrcode/qrcode.go` (exported funcs) |
| 13 | **Config features** | AGENTS.md §Config table | `internal/config/config.go` (Env detection, SecurePath) |
| 14 | **Version string** | AGENTS.md header + footer | `internal/config/config.go` (Version const) |
| 15 | **Documentation files** | AGENTS.md §Documentation Reference | `documentation/` directory listing |

### Phase 3: Fix discrepancies

For each discrepancy found:

1. **Classify severity**:
   - CRITICAL: Documents non-existent type/function, wrong method receiver
   - MAJOR: Missing entity types, wrong environment detection, wrong file paths
   - MINOR: Missing methods, imprecise descriptions, wrong parameter names

2. **Edit docs** (prefer fixing docs over code unless code is buggy):
   - Use `edit` tool with exact `oldString`/`newString`
   - Preserve markdown table formatting
   - Keep descriptions concise

3. **If code has a real bug** (e.g. Makefile LDFLAGS wrong path):
   - Fix the code too
   - Note it in the summary

### Phase 4: Verify

1. Run `go build ./cmd/tamk` to ensure code compiles
2. Run `go test -short ./...` if time permits
3. Run `.agents/hooks/work-finished`

### Phase 5: Report

Output a table:

```
| # | Severity | File | Fix |
|:--|:--|:--|:--|
| 1 | CRITICAL | AGENTS.md:458 | VersionInfo → UpdateInfo |
| 2 | MAJOR | AGENTS.md:579 | Added 8 missing error sentinels |
```

## Quick checklist for subagents

When spawning an explore agent for T2T, use this prompt template:

```
Cross-verify AGENTS.md and documentation/ against actual code in TAMK.
Check these 15 areas: file structure, CLI commands, entity types,
use case methods, repository interfaces, template files, error sentinels,
watcher extensions/ignore dirs, logger functions, QR code functions,
config features, version string, documentation files.

For each discrepancy report: file path + line, what docs say vs what
code does, severity (CRITICAL/MAJOR/MINOR).
```

## Notes

- AGENTS.md is the primary docs source (700+ lines, 16 sections)
- documentation/ has 21 .md files + VERSIONING.txt
- Entity types total 11: Project, ProjectType, WebContentMode, BuildResult, BuildPhase, BuildCache, Template, TemplateMapping, Keystore, UpdateInfo, UpdateLevel
- Error sentinels total 15 + BuildError struct
- Watch extensions: .html .css .js .json .png .jpg .jpeg .svg .webp .xml .kt
- Ignore dirs: .git node_modules secret .idea
