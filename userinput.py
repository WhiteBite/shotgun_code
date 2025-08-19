\
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
