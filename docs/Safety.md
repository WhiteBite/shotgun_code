# Safety & Security

- Sandbox: Podman rootless, сеть off по умолчанию; cgroups CPU/RAM/timeout.
- Path sandbox: запрет абсолютных путей и выхода за workspace.
- Белые списки команд для process-исполнения.
- Redaction: чистка логов и контекстов от секретов (regex+entropy+allowlist).
- SBOM/лицензии: Syft/Grype; Cosign подписи (в перспект.).
- Governance: PR-шаблоны, codeowners, require review для high-risk зон (auth/crypto/infra).
