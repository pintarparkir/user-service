# Architecture Diagrams — user-service

| File | Purpose | Audience |
|------|---------|----------|
| `erd.svg` | Entity Relationship Diagram (user_profile, vehicle, idempotency_key) | Backend · DBA |

System-level diagrams (HLD, sequence, state machine) live as ASCII / Mermaid blocks
inside the relevant docs (`high-level-design.md`, `../runbook/demo-walkthrough.md`)
so they stay in sync with the code.

## Editing the ERD

1. Open `erd.svg` in any text editor or in Figma / draw.io.
2. Re-export as SVG when done.
3. Mermaid source-of-truth lives in `../erd.md` — keep them aligned.
