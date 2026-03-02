# GAIOL Repository Reorganization Plan

This document records the reorganization applied to keep the repo professional and navigable.

## Principles

- **Root**: Only entry points (README, QUICKSTART, API.md, LICENSE), env template, and start/stop scripts.
- **Docs**: All other documentation under `docs/` with a single index at `docs/README.md`.
- **Migrations**: Canonical run order in `migrations/README.md`; 001 available as single file or chunks.
- **Scripts**: Test and dev scripts under `scripts/` with lowercase-with-dashes names and a README.
- **Artifacts**: Build and test output files ignored via `.gitignore`; `make clean` removes them.

## Steps Executed

1. **migrations/README.md** — Added with two tracks: core (001 → 007 → 008) and optional (002, 003, 004, 006); chunk alternative for 001 documented.
2. **.gitignore** — Added `build_error.txt`, `test_results.txt`, `TEST_LOG.txt`, `test_errors.txt`, `reasoning.a`.
3. **scripts/** — Created `scripts/README.md`; moved test/dev scripts into `scripts/test/` and `scripts/dev/` with consistent names.
4. **docs/** — Moved root-level guides and references into `docs/` with lowercase-with-dashes filenames; added `docs/README.md` as the documentation index; kept existing `docs/` filenames (RUNBOOK.md, IMPLEMENTATION_PHASES.md, etc.) for stability.
5. **_archive/README.md** — Added to state that `_archive` is legacy/superseded content.
6. **Makefile** — Extended `clean` to remove common artifact files.
7. **README.md / QUICKSTART.md** — Updated links to point to `docs/database-setup.md` and `docs/README.md` where appropriate.
8. **Cross-references** — Updated RUNBOOK and other docs to use new paths under `docs/`.

## File Mapping (Root → docs/)

| Was (root) | Now |
|------------|-----|
| DATABASE_SETUP.md | docs/database-setup.md |
| AUTHENTICATION.md | docs/authentication.md |
| ARCHITECTURE.md | docs/architecture.md |
| ROUTING.md | docs/routing.md |
| OLLAMA_SETUP.md | docs/ollama-setup.md |
| SIMPLIFIED_ARCHITECTURE.md | docs/simplified-architecture.md |
| FEATURES_IMPLEMENTED.md | docs/features-implemented.md |
| IMPLEMENTATION_STATUS.md | docs/implementation-status.md |
| DESIGN_ACTION_PLAN.md | docs/design-action-plan.md |
| DESIGN_ENHANCEMENT_GUIDE.md | docs/design-enhancement-guide.md |
| QUICK_DESIGN_WINS.md | docs/quick-design-wins.md |
| COMPARISON.md | docs/comparison.md |
| WORLD_MODEL_*.md | docs/world-model-*.md (implementation, quick-start, verification) |
| PROJECT_DOCUMENTATION.md | docs/project-documentation.md |
| DOCUMENTATION.md | Replaced by docs/README.md (index) |
| CLEANUP_SUMMARY.md, CODEBASE_REVIEW_SUMMARY.md, FAVICON_NOTE.md | docs/ (reference/archive) |

## Script Mapping (Root → scripts/)

| Was (root) | Now |
|------------|-----|
| test-integration.ps1 | scripts/test/integration.ps1 |
| QUICK-TEST.ps1 | scripts/test/quick.ps1 |
| FINAL-TEST.ps1 | scripts/test/final.ps1 |
| RAW-TEST.ps1 | scripts/test/raw.ps1 |
| test-pipeline.ps1 | scripts/test/pipeline.ps1 |
| TEST-OLLAMA.ps1 | scripts/test/ollama.ps1 |
| clean-start.ps1 | scripts/dev/clean-start.ps1 |

Start/stop remain at root: start.ps1, stop.ps1, start.sh, start.bat, stop.bat.
