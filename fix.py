#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
fix.py — добавить MCP-правила и "умный" userinput.py (не трогает cutc.mdc).

Флаги:
  --force           перезаписать целевые файлы
  --dry-run         показать изменения без записи
  --patch-taskflow  дописать MCP-блок в .cursor/rules/70_taskflow.mdc (если нет маркера)
  --gate-wait       обновить userinput.py с режимами ожидания (ALWAYS|OFF|AUTO|ONCE)
"""
import argparse
from pathlib import Path
from textwrap import dedent
import json

ROOT = Path(__file__).resolve().parent
RULE_80 = ROOT / ".cursor" / "rules" / "80_mcp_requirements.mdc"
RULE_85 = ROOT / ".cursor" / "rules" / "85_memory_policy.mdc"
TASKFLOW = ROOT / ".cursor" / "rules" / "70_taskflow.mdc"
USERINPUT = ROOT / "userinput.py"

BLOCK_TASKFLOW = dedent("""\
\n# MCP usage and acceptance (mandatory)
- Все операции с кодом/файлами/процессами выполнять ТОЛЬКО через MCP‑инструменты:
  - filesystem: чтение/запись/листинг в пределах workspace (sandbox).
  - git: статус/коммиты/ветки/дифф (worktree).
  - process: только из белого списка (go, npm, node, tsc, eslint, prettier, rg/grep) с таймаутами и лимитами.
  - search: ripgrep/ctags для навигации.
- Любая сеть/облачный LLM — ТОЛЬКО через requiresEscalation: true, с минимизацией данных (Context Pack) и логированием provenance.
- Любые правки — ТОЛЬКО Edits JSON/SAP, без git diff и без свободного текста.
""")

CONTENT_80 = dedent("""\
---
alwaysApply: true
priority: high
tags: ['mcp','safety','policy']
---
# MCP Requirements

Allowed tools:
- filesystem: path sandbox = workspace root; запрещены абсолютные пути и ../
- git: status/diff/commit/branch/worktree; push — по политике
- process: allowlist = ["go","git","npm","node","tsc","eslint","prettier","rg","grep"], таймауты/лимиты обязательны
- search: ripgrep/ctags

Forbid:
- любые команды вне allowlist
- изменение go.mod, package.json/lock, dist/node_modules, secrets
- чтение/запись вне workspace

Escalation:
- сеть/облако/внешние API = requiresEscalation: true + data minimization (Context Pack), логирование provenance.
""")

CONTENT_85 = dedent("""\
---
alwaysApply: true
tags: ['memory','policy']
---
# Memory Usage Policy (MCP memory/Byteover)

- Хранить ТОЛЬКО метаданные: current_task_id, iteration, last_error_code/summary, timestamps.
- НЕ хранить исходники, секреты, большие логи.
- Истина — в репозитории: tasks/status.json, build/test reports.
- Очистка: при завершении task — сворачивать память до краткого summary.
""")

# ВАЖНО: внешняя строка — тройные одинарные кавычки, чтобы внутри можно было """ (docstring)
GATED_USERINPUT = dedent(r'''\
#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Gated userinput.py
Режимы: ALWAYS | OFF | AUTO | ONCE
- Источник: env WAIT_MODE или .cursor_wait.json {"mode":"AUTO","flagPath":".wait.flag"}
- AUTO: ждать, если есть флаг-файл или tasks/status.json содержит requireHuman=true
Совместим с прежним next_prompt.txt.
"""
import os, sys, time, json
from pathlib import Path

PROMPT_FILE = Path('next_prompt.txt')
ANSI_CLEAR_SCREEN = '\033[2J\033[H'
ANSI_CLEAR_LINE   = '\033[2K\033[1G'
PULSE = ['⣷','⣯','⣟','⡿','⢿','⣻','⣽','⣾']

def load_mode():
    mode = os.getenv("WAIT_MODE", "").strip().upper()
    if mode in {"ALWAYS","OFF","AUTO","ONCE"}:
        return mode, ".wait.flag"
    cfg_file = Path(".cursor_wait.json")
    if cfg_file.exists():
        try:
            cfg = json.loads(cfg_file.read_text(encoding="utf-8"))
            m = str(cfg.get("mode","ALWAYS")).upper()
            f = str(cfg.get("flagPath",".wait.flag"))
            if m in {"ALWAYS","OFF","AUTO","ONCE"}:
                return m, f
        except Exception:
            pass
    return "ALWAYS", ".wait.flag"

def should_wait(mode, flag_path):
    if mode == "OFF":
        return False
    if mode in {"ALWAYS","ONCE"}:
        return True
    # AUTO
    if flag_path and Path(flag_path).exists():
        return True
    st = Path("tasks/status.json")
    if st.exists():
        try:
            data = json.loads(st.read_text(encoding="utf-8"))
            if any(t.get("requireHuman") for t in data.get("tasks",[])):
                return True
        except Exception:
            pass
    return False

def clear_console():
    sys.stdout.write(ANSI_CLEAR_SCREEN)
    sys.stdout.flush()

def wait_for_prompt():
    pulse_idx = 0
    while not PROMPT_FILE.exists() or PROMPT_FILE.stat().st_size == 0:
        spinner = PULSE[pulse_idx % len(PULSE)]
        sys.stdout.write(f"{ANSI_CLEAR_LINE}{spinner} Ожидание команды...")
        sys.stdout.flush()
        pulse_idx += 1
        time.sleep(0.1)
    text = PROMPT_FILE.read_text(encoding='utf-8').strip()
    try:
        PROMPT_FILE.unlink()
    except Exception:
        pass
    return text

def main():
    mode, flag_path = load_mode()
    if not should_wait(mode, flag_path):
        print("Skip waiting (WAIT_MODE={}): proceeding automatically.".format(mode))
        sys.exit(0)
    try:
        clear_console()
        cmd = wait_for_prompt()
        clear_console()
        print(cmd)
        if mode == "ONCE":
            cfg = {"mode":"OFF","flagPath": flag_path or ".wait.flag"}
            Path(".cursor_wait.json").write_text(json.dumps(cfg, indent=2, ensure_ascii=False), encoding="utf-8")
        sys.exit(0)
    except KeyboardInterrupt:
        sys.stdout.write(f"\n{ANSI_CLEAR_LINE}Отменено.\n")
        sys.exit(1)
    except Exception as e:
        print(f"Критическая ошибка: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
''')

def write_file(p: Path, content: str, force: bool, dry: bool):
    p.parent.mkdir(parents=True, exist_ok=True)
    if p.exists() and not force:
        return False, f"skip (exists) {p}"
    if not dry:
        p.write_text(content, encoding="utf-8")
    return True, f"write {p}"

def patch_taskflow(taskflow: Path, dry: bool):
    if not taskflow.exists():
        return False, f"skip (not found) {taskflow}"
    txt = taskflow.read_text(encoding="utf-8")
    marker = "MCP usage and acceptance (mandatory)"
    if marker in txt:
        return False, f"skip (already contains MCP block) {taskflow}"
    if not dry:
        taskflow.write_text(txt.rstrip() + BLOCK_TASKFLOW, encoding="utf-8")
    return True, f"append MCP block to {taskflow}"

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--force", action="store_true")
    ap.add_argument("--dry-run", action="store_true")
    ap.add_argument("--patch-taskflow", action="store_true")
    ap.add_argument("--gate-wait", action="store_true")
    args = ap.parse_args()

    actions = []
    ok, msg = write_file(RULE_80, CONTENT_80, args.force, args.dry_run); actions.append(msg)
    ok, msg = write_file(RULE_85, CONTENT_85, args.force, args.dry_run); actions.append(msg)

    if args.patch_taskflow:
        ok, msg = patch_taskflow(TASKFLOW, args.dry_run); actions.append(msg)
    else:
        actions.append(f"info: not patching {TASKFLOW} (use --patch-taskflow)")

    if args.gate_wait:
        ok, msg = write_file(USERINPUT, GATED_USERINPUT, args.force, args.dry_run); actions.append(msg)
        cfg_path = ROOT / ".cursor_wait.json"
        if not args.dry_run:
            if not cfg_path.exists():
                cfg = {"mode":"ALWAYS","flagPath":".wait.flag"}
                cfg_path.write_text(json.dumps(cfg, indent=2, ensure_ascii=False), encoding="utf-8")
                actions.append(f"write {cfg_path}")
            else:
                actions.append(f"skip (exists) {cfg_path}")
    else:
        actions.append("info: userinput.py not changed (use --gate-wait)")

    print("=== fix.py report ===")
    for a in actions:
        print(" -", a)
    print("OK.")

if __name__ == "__main__":
    main()