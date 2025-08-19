# Shotgun — Архитектура

Цель: автономный AI‑код‑агент. Собирает минимальный контекст, генерирует структурированные правки (Edits JSON / SAP),
применяет их семантически (AST/LSP/anchors), локально верифицирует (build/tests/static/mutation/licensing) и только затем
предлагает PR/автозаливку по политике (gate).

Слои:
- UI (Wails): ввод задач, derived diff и отчёты, кнопки Apply/Revert/Repair/PR.
- Оркестратор (Go): координация Retrieval → Context Pack → LLM (edits) → Apply → Verify → Gate → Output.
- Indexer/Code Intelligence: LSP (gopls/tsserver/jdt.ls/pyright/rust-analyzer), LSIF/SCIP, Tree-sitter, Zoekt/Lucene, (Qdrant/Faiss — опционно).
- Context Builder: Impact-slicing, «умные вырезки», стабильная упаковка.
- Patch Synthesizer (LLM): локальные модели по умолчанию; облако — по политике.
- Apply Engine: ast/symbol → anchor(+hash) → fullFile, форматирование и импорт-фиксы, derived diff.
- Sandbox/Runner: Podman rootless, сеть off, cgroups, git worktree/overlay.
- Verify & Gate: build/type-check → targeted tests → static analysis → mutation testing; SBOM, лицензии, секьюрити.
- Repair Loop: rule-fix → retry.
- Output & Governance: отчёты, PR, подписи (Cosign), аудиты, метрики.

Схемы/контракты см.:
- docs/LLM_Contracts.md (Edits JSON / SAP, Context Pack, GBNF)
- docs/Retrieval_Policy.md
- docs/Verification_Pipeline.md
- docs/Safety.md
- docs/Runbook.md
