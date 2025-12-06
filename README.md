# Shotgun Code

[Русская версия](README.ru.md)

Desktop application for building code context for AI assistants. Collect, filter, and export your codebase in formats optimized for LLM consumption.

## Features

### Project Management

- Open local folders or clone from GitHub/GitLab
- Recent projects list with quick access
- Real-time file system watching

### File Explorer

- Tree view with expand/collapse
- Multi-select with Ctrl/Shift
- Quick filters: by extension, git status, size
- Ignore rules: .gitignore support + custom patterns
- File preview with syntax highlighting

### Context Building

- Select files to include in context
- Token counting with model limits display
- Multiple output formats: Markdown, XML, JSON, Plain text
- Manifest generation with file tree
- Comment stripping option

### Export Options

- Copy to clipboard
- Export to PDF (for AI with file upload)
- Split large contexts into chunks
- Context history with restore

### AI Integration

- Multiple providers: OpenAI, Gemini, OpenRouter, LocalAI, Qwen
- Model selection per provider
- Response caching for identical requests
- Streaming responses

### Git Integration

- Branch switching
- File status indicators (modified, added, deleted)
- Filter by git status
- GitHub/GitLab repository cloning

## Installation

Download the latest release for your platform:

- Windows: `shotgun-code-windows-amd64.exe`
- macOS: `shotgun-code.app.zip`
- Linux: `shotgun-code-linux-amd64`

Or build from source (requires Go 1.24+, Node.js 20+):

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Clone and build
git clone https://github.com/WhiteBite/shotgun_code.git
cd shotgun_code
wails build
```

## Usage

1. Open a project folder or clone a repository
2. Select files in the explorer (use filters to narrow down)
3. Review token count and adjust selection
4. Copy context to clipboard or export to PDF
5. Paste into your AI assistant

## Tech Stack

- Backend: Go, Wails
- Frontend: Vue 3, TypeScript, Pinia, Tailwind CSS
- Build: Vite, GitHub Actions

## License

MIT
