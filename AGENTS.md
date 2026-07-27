# AGENTS.md — Instructions for AI Coding Agents

> **Operational Contract & Guidelines for AI Assistant Co-Developers**
>
> All AI agents working on the `fluint` codebase MUST strictly comply with this document.

---

## 1. Project Overview & Core Requirements

**Fluint** is an open-source, high-performance Terminal User Interface (TUI) engine for Go built **from scratch**.

- **Key Focus:** Smooth discrete cell animations, advanced VFX (easing, sub-cell rendering, transitions), and an integrated UI kit.
- **Forbidden Core Dependencies:** `Bubble Tea`, `tcell`, `termbox`, `tview`, `lipgloss`, or any high-level TUI frameworks.
- **Allowed Core Dependencies:** `golang.org/x/sys` (termios/ioctl). Any additional dependency requires explicit user approval and a written rationale.
- **Autonomy Level:** **LOW**. Agents propose architectural solutions with justifications but do not adopt them unilaterally. Public API changes, new dependencies, structural changes, and architectural forks require user approval.

---

## 2. Documentation Directory & Architectural Records

Documentation uses a **mirrored 3-language structure** under [`/docs`](file:///e:/pet/fluint/docs):

- **English Docs:** [`/docs/en`](file:///e:/pet/fluint/docs/en) (Base/Default)
- **Russian Docs:** [`/docs/ru`](file:///e:/pet/fluint/docs/ru)
- **Chinese Docs:** [`/docs/zh`](file:///e:/pet/fluint/docs/zh)
- **Architecture Decision Records (ADR):**
  - English: [`/docs/en/adr`](file:///e:/pet/fluint/docs/en/adr)
  - Russian: [`/docs/ru/adr`](file:///e:/pet/fluint/docs/ru/adr)
  - Chinese: [`/docs/zh/adr`](file:///e:/pet/fluint/docs/zh/adr)
- **Architecture Overview:** [`ARCHITECTURE.md`](file:///e:/pet/fluint/ARCHITECTURE.md) → [`docs/en/ARCHITECTURE.md`](file:///e:/pet/fluint/docs/en/ARCHITECTURE.md)

---

## 3. Git, Branching, Commit, and PR Conventions

1. **Branching Strategy:**
   - Every feature, bugfix, CI update, or infrastructure task **MUST** be implemented in its own dedicated branch (e.g., `feature/<name>`, `fix/<name>`, `ci/<name>`, `infra/<name>`).
   - All topic branches are merged **exclusively into `dev`**. Direct commits to `main` or `dev` without a topic branch are prohibited.
2. **Commit Language & Messages:**
   - **ALL commit messages MUST be in English.**
   - Follow Conventional Commits format (e.g., `feat(term): add zero-alloc escape parser`, `fix(buffer): resolve race condition in diff`).
3. **Automated PR Creation:**
   - After the USER approves a completed task, the agent MUST create a Pull Request from `dev` to `main` using `gh pr create --base main --head dev --title "..." --body "..."`. The agent MUST ONLY create the PR and NEVER merge it.

---

## 4. Multilingual Documentation Policy

- **Mirrored Languages Structure:** All project documentation MUST be maintained with an identical directory tree across three language folders:
  1. **English:** [`docs/en/...`](file:///e:/pet/fluint/docs/en)
  2. **Russian:** [`docs/ru/...`](file:///e:/pet/fluint/docs/ru)
  3. **Chinese:** [`docs/zh/...`](file:///e:/pet/fluint/docs/zh)
- **Root Files Cleanliness:** Do NOT create `.ru.md` or `.zh.md` files in the root folder. Root files (`README.md`, `ARCHITECTURE.md`, `CHANGELOG.md`) remain English-focused with top navigation bars pointing to language subdirectories.
- **Synchronization Rule:** After creating or modifying any feature, ADR, guide, or architectural doc, the agent **MUST synchronize all documentation across all three language directories (`docs/en/`, `docs/ru/`, `docs/zh/`)** before ending the turn.

---

## 5. Versioning and CHANGELOG

- **Semantic Versioning:** Follow SemVer (`vX.Y.Z`).
- **CHANGELOG Maintenance:** Update [`docs/en/CHANGELOG.md`](file:///e:/pet/fluint/docs/en/CHANGELOG.md), [`docs/ru/CHANGELOG.md`](file:///e:/pet/fluint/docs/ru/CHANGELOG.md), and [`docs/zh/CHANGELOG.md`](file:///e:/pet/fluint/docs/zh/CHANGELOG.md) for every new version release.
- **No "Unreleased" Headers:** NEVER use "Unreleased", "Невыпущенное", or "未发布" in changelogs. Always write the explicit next target version number following the previous release (e.g., `## [v0.1.1]` or `## [v0.2.0]`).
- **Git Tagging:** Create a new git tag (e.g., `git tag -a v0.1.0 -m "Release v0.1.0"`) for every version bump.

---

## 6. Required Skills Usage Workflow

Agents must actively invoke available skills for specific tasks:

1. **Documentation Writing:**
   - **Skill:** `documentation-writer`
   - Apply Diátaxis framework principles (Tutorials, How-to Guides, Reference, Explanation).
2. **Go Development & Review:**
   - **Skills:** `golang-code-style`, `golang-design-patterns`, `golang-error-handling`, `golang-performance`, `golang-security`.
   - Adhere strictly to Go idioms, error wrapping (`%w`), zero-allocation guidelines, and functional options.
3. **Security Audits (Mandatory Pre-Feature Step):**
   - **Skills:** `security-review`, `security-best-practices`.
   - **Requirement:** **BEFORE** starting any new feature implementation, run a security audit using `security*` skills to inspect existing and planned code for vulnerabilities.

---

## 7. Performance, Benchmarking & Fuzzing Doctrine

1. **Steady-State Target:** **0 allocations per frame** in hot path (`diff`, `render`, `event dispatch`, `anim tick`).
2. **I/O Goal:** Exactly **one `write(2)` syscall per frame**.
3. **Pre/Post Measurement Protocol:**
   - **BEFORE** introducing changes to hot-path code: run benchmarks and fuzzing tests, saving baseline results (`old.txt`).
   - **AFTER** making changes: re-run benchmarks and fuzz tests (`new.txt`).
4. **Performance Regression Rule:**
   - If performance degrades noticeably (>10% slowdown or increased allocations), the agent **IS OBLIGATED** to investigate the root cause, identify bottlenecks (via `pprof`/`benchstat`), and apply optimizations until performance is restored or improved.

---

## 8. Checklist Before Submitting Code

- [ ] Security review performed via `security*` skills.
- [ ] Code compiles cleanly across targets (`go build ./...`).
- [ ] No prohibited framework imports (§1).
- [ ] All tests pass with race detector enabled (`go test -race ./...`).
- [ ] Benchmarks & fuzzing executed pre- and post-edit; performance validated via `benchstat`.
- [ ] Zero allocations in hot path verified (`testing.AllocsPerRun`).
- [ ] Documentation updated and synchronized across `docs/en/`, `docs/ru/`, and `docs/zh/`.
- [ ] Git commit messages written in English.
