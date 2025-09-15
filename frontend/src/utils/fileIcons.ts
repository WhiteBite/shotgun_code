export function getFileIcon(name: string): string {
  const ext = name.split(".").pop()?.toLowerCase() || "";
  const map: Record<string, string> = {
    js: "🟨",
    ts: "🔷",
    json: "📋",
    md: "📝",
    vue: "💚",
    go: "💧",
    py: "🐍",
    html: "🌐",
    css: "🎨",
    scss: "🎨",
    dockerfile: "🐳",
    yml: "📋",
    yaml: "📋",
    gitignore: "🚫",
    ps1: "📜",
    sh: "📜",
  };
  return map[ext] || "📄";
}