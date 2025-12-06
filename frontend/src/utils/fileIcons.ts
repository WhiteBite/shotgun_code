export interface FileTypeInfo {
  icon: string
  colorClass: string
}

// Единый источник иконок и цветов для типов файлов
export const FILE_TYPE_CONFIG: Record<string, FileTypeInfo> = {
  ts: { icon: '🔷', colorClass: 'bg-blue-500' },
  tsx: { icon: '⚛️', colorClass: 'bg-cyan-500' },
  js: { icon: '🟨', colorClass: 'bg-yellow-500' },
  jsx: { icon: '⚛️', colorClass: 'bg-cyan-500' },
  vue: { icon: '💚', colorClass: 'bg-emerald-500' },
  go: { icon: '💧', colorClass: 'bg-sky-500' },
  py: { icon: '🐍', colorClass: 'bg-yellow-600' },
  css: { icon: '🎨', colorClass: 'bg-pink-500' },
  scss: { icon: '🎨', colorClass: 'bg-pink-600' },
  sass: { icon: '🎨', colorClass: 'bg-pink-600' },
  html: { icon: '🌐', colorClass: 'bg-orange-500' },
  json: { icon: '📋', colorClass: 'bg-gray-500' },
  md: { icon: '📝', colorClass: 'bg-gray-400' },
  yaml: { icon: '📋', colorClass: 'bg-red-400' },
  yml: { icon: '📋', colorClass: 'bg-red-400' },
  sql: { icon: '🗃️', colorClass: 'bg-indigo-500' },
  dockerfile: { icon: '🐳', colorClass: 'bg-blue-400' },
  gitignore: { icon: '🚫', colorClass: 'bg-gray-500' },
  ps1: { icon: '📜', colorClass: 'bg-blue-300' },
  sh: { icon: '📜', colorClass: 'bg-green-500' },
  java: { icon: '☕', colorClass: 'bg-orange-600' },
  cpp: { icon: '⚙️', colorClass: 'bg-blue-600' },
  c: { icon: '⚙️', colorClass: 'bg-blue-500' },
  rs: { icon: '🦀', colorClass: 'bg-orange-500' },
  default: { icon: '📄', colorClass: 'bg-gray-500' }
}

export function getFileIcon(name: string): string {
  const ext = name.split('.').pop()?.toLowerCase() || ''
  return FILE_TYPE_CONFIG[ext]?.icon || FILE_TYPE_CONFIG.default.icon
}

export function getFileTypeInfo(name: string): FileTypeInfo {
  const ext = name.split('.').pop()?.toLowerCase() || ''
  return FILE_TYPE_CONFIG[ext] || FILE_TYPE_CONFIG.default
}