# Conventions

- Именование путей/файлов — в стиле проекта; монорепы — с префиксами пакетов (packages/*, services/*).
- Go: gofmt/goimports — обязательны; ошибки с контекстом fmt.Errorf("...: %w").
- TS/Vue: строгий TS, Prettier + ESLint --fix; корректная регистрация компонентов/роутов/stores.
- Java/Kotlin: OpenRewrite для семантических правок; Gradle/Maven pipeline.
- Python: LibCST для безопасных правок, black/ruff/mypy.
- Rust: cargo fmt + clippy; AST с syn/quote.
- Terraform/YAML: HCL2 AST и tflint; policy as code (OPA/Rego).
