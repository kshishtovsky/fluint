# ADR-0004: MIT License

## Status
Accepted

## Context
The project ships source under an explicit licence. The choice affects
every downstream consumer: permissive licences encourage adoption,
restrictive licences preserve commercial optionality.

We evaluated MIT, BSD-3-Clause, Apache-2.0, and GPL-3.0.

## Decision
**MIT License** — Copyright (c) 2026 Vitaly &lt;mihaylovvsjob@gmail.com&gt;.

The full text is in `LICENSE.md` (English) and mirrored in
`docs/ru/LICENSE.md` and `docs/zh/LICENSE.md` for non-English readers.
The English version is the legally binding one; the translations are
informational only.

## Rationale

- MIT is the most widely accepted permissive licence in the Go
  ecosystem and minimises friction for downstream users.
- The two-paragraph text is short, unambiguous, and easy to verify
  against existing tools (GitHub licence detection, licensecheck).
- Aligns with the project's open-source, from-scratch philosophy
  (AGENTS.md §1) — maximise reuse, minimise adoption friction.

## Consequences

- No copyleft obligations for downstream users.
- The project ships without a patent grant (MIT does not include
  one). If patent protection becomes a concern, we can migrate to
  Apache-2.0 in a future ADR.
- The legal text is the English `LICENSE.md`; translations are
  reference only and cannot override the English version in case of
  conflict.