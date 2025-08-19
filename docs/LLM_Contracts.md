# LLM Contracts — Edits JSON / SAP и Context Pack

Базовый формат — Edits JSON / SAP: единый «контейнер» для разных операций:
- recipeOp (OpenRewrite/ts-morph/go-ast/LibCST)
- workspaceEdit (LSP-совместимые диапазонные правки)
- astOp (языко-специфичные операции: insertFunc/addImport/updateRoute)
- anchorPatch (якорные вставки: before/after + hash)
- fullFile (create/replace/delete)

## Edits JSON Schema
См. docs/schemas/edits.schema.json

## GBNF Grammar (llama.cpp)
См. docs/schemas/edits.gbnf

## Context Pack Schema
См. docs/schemas/context_pack.schema.json

## Пример (минимальный)
{
  "schemaVersion": "1.0",
  "edits": [
    {
      "kind": "recipeOp",
      "path": "backend/application/ai_service.go",
      "language": "go",
      "recipe": { "engine": "go-ast", "action": "insertFunc", "params": { "receiver": "AIService", "name": "NewMethod" } },
      "post": { "formatters": ["gofmt","goimports"] }
    }
  ]
}
