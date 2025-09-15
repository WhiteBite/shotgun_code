import { container } from "@/infrastructure/container";
import type { ProjectRepository } from '@/domain/repositories/ProjectRepository';
import { APP_CONFIG } from '@/config/app-config';

export type QuickLookType = "fs" | "git" | "text";

export interface QuickLookData {
  rootDir: string;
  path: string;
  type: QuickLookType;
  commitHash?: string;
  event: MouseEvent | null;
  isPinned?: boolean;
  position?: { x: number; y: number };
  content?: string; // Для типа "text"
}

export interface QuickLookState {
  isActive: boolean;
  path: string;
  content: string;
  language: string;
  truncated: boolean;
  error: string | null;
  isPinned: boolean;
  position: { x: number; y: number } | null;
}

export function detectLanguage(filePath: string): string {
  const ext = filePath.split(".").pop()?.toLowerCase();
  const languageMap: Record<string, string> = {
    js: "javascript",
    ts: "typescript",
    jsx: "javascript",
    tsx: "typescript",
    vue: "vue",
    html: "html",
    css: "css",
    scss: "scss",
    sass: "sass",
    less: "less",
    json: "json",
    xml: "xml",
    yaml: "yaml",
    yml: "yaml",
    md: "markdown",
    py: "python",
    java: "java",
    c: "c",
    cpp: "cpp",
    cs: "csharp",
    php: "php",
    rb: "ruby",
    go: "go",
    rs: "rust",
    swift: "swift",
    kt: "kotlin",
    scala: "scala",
    sql: "sql",
    sh: "bash",
    ps1: "powershell",
    bat: "batch",
    dockerfile: "dockerfile",
    gitignore: "gitignore",
    env: "properties",
    toml: "toml",
    ini: "ini",
    conf: "conf",
    log: "log",
    txt: "text",
    text: "text",
  };

  return languageMap[ext || ""] || "text";
}

export function computePanelPos(event?: MouseEvent) {
  const safeTop = (y: number) =>
    Math.max(8, Math.min(y, window.innerHeight - 160));
  const safeLeft = (x: number) =>
    Math.max(8, Math.min(x, window.innerWidth - 620));
  if (!event) {
    return {
      top: safeTop(120) + "px",
      left: safeLeft(window.innerWidth / 2 - 300) + "px",
    };
  }
  return {
    top: safeTop(event.clientY) + "px",
    left: safeLeft(event.clientX) + "px",
  };
}

export async function loadAndHighlight(
  projectRepository: ProjectRepository,
  rootDir: string,
  path: string,
  type: QuickLookType,
  commitHash?: string,
  content?: string,
): Promise<{ content: string; language: string; truncated: boolean }> {
  try {
    if (type === "text" && content) {
      // Для текстового типа возвращаем переданное содержимое
      return {
        content: content,
        language: "text",
        truncated: false,
      };
    }

    let raw = "";
    if (type === "fs") {
      // Use the ProjectRepository directly
      raw = await projectRepository.readFileContent(rootDir, path);
    } else if (type === "git") {
      if (!commitHash)
        throw new Error("commitHash is required for git preview");
      // Use the Wails API directly instead of apiService
      raw = await window.go.main.App.GetFileContentAtCommit(rootDir, path, commitHash);
    }

    // Ограничиваем размер для больших файлов
    const maxChars = APP_CONFIG.ui.quicklook.MAX_CHARS;
    let truncated = false;
    if (raw.length > maxChars) {
      raw = raw.slice(0, maxChars);
      truncated = true;
    }

    const lang = detectLanguage(path);

    return {
      content: raw,
      language: lang,
      truncated,
    };
  } catch (error) {
    console.error("QuickLook error:", error);
    throw error;
  }
}