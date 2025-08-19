# MCP (Model Context Protocol) для Shotgun

Цель MCP — дать LLM безопасные «органы» для операций: контекст, правки, форматирование, сборка/тесты, статанализ, лицензии, git/PR.
Рекомендуемые инструменты:
- context.suggest({prompt,target?}) -> {files[], symbols[], reasons[]}
- context.build({files|symbols}) -> {manifest, blocks[], tokens, contextPack}
- edits.apply({editsJson}) -> {applied, changedFiles[], diagnostics[], derivedDiff}
- code.format({paths|all}) -> {ok, logs}
- build.check({scope}) -> {ok, output}
- test.run({scope}) -> {ok, summary}
- analyze.static({scope}) -> {issues[]}
- license.scan({scope}) -> {sbom, licenses[], violations[]}
- secrets.scan({scope}) -> {findings[]}
- coverage.map({scope}) -> {file->lines, test->files}
- git.stageAndCommit({message}) -> {ok, commitId}
- info.limits() -> {maxFileSize, maxTokens, timeouts}

Безопасность:
- Любая внешняя сеть/облачный LLM — requiresEscalation: true, с журналированием и минимизацией данных.
- Path sandbox, timeouts, quotas; белые списки команд.

Пример конфигурации MCP-серверов (псевдоконфиг — адаптируйте под ваш рантайм):
{
  "mcpServers": {
    "filesystem": { "command": "mcp-filesystem", "args": ["--root","${workspaceFolder}"] },
    "process":    { "command": "mcp-process",    "args": ["--allow","go,git,npm,gradle,mvn,tsc"] },
    "git":        { "command": "mcp-git" },
    "search":     { "command": "mcp-search",     "args": ["--rg"] },
    "shotgun":    { "command": "shotgun-mcp",    "args": [] }
  }
}

Полезно: каталог MCP‑серверов и практики — см. https://cursor.directory/mcp
(Выберите совместимые реализации и сверьте параметры/безопасность под свою среду.)
