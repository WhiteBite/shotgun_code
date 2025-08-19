# Verification Pipeline

Порядок:
1) format/import-fix
2) build/type-check (go build/staticcheck; tsc/ESLint; javac/Gradle; mypy/ruff; cargo check)
3) tests: targeted (affected graph) + smoke; rerun фейлов N=3 (flaky quarantine)
4) static security: Semgrep + языкоспецифичные анализаторы
5) mutation testing: Stryker (JS/TS) / PITest (Java) — порог ≥ 60–70% для automerge
6) SBOM/licensing: Syft + Grype; policy stop для запрещённых лицензий/уязвимостей
7) Policy as code: OPA/Rego — глобальные запреты/бюджеты
