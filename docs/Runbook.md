# Runbook

- Dev:
  - Индексация: поднимаются LSP-сервера (gopls/tsserver/jdt.ls/pyright/rust-analyzer).
  - Сбор контекста: Retrieval Policy → Context Pack.
  - Применение: apply engine (ast→anchor→fullFile) → format/imports → derived diff.
  - Проверка: build/type-check → targeted tests → static → (mutation) → SBOM/licensing → OPA.
- Траблшутинг:
  - Диагностика с кодами ошибок (E-AST-PARSE, E-ANCHOR-NOT-FOUND, E-BUILD-FAIL, ...).
  - Если нужен облачный LLM → requiresEscalation: true + Context Pack (минимум кода).
- Метрики:
  - time_to_green (p50/p95), iterations_to_green, mutation_score, auto_merge_rate, revert_rate, security_violation_rate, llm_escalation_rate.
