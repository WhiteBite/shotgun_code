// Syntax highlighting utilities for context preview

export function escapeHtml(text: string): string {
    return text
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
}

export function escapeRegex(text: string): string {
    return text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

export function highlightXml(line: string): string {
    return line
        .replace(/(&lt;\/?)(file|content)(&gt;)/g, '<span class="syntax-tag">$1$2$3</span>')
        .replace(/(&lt;file\s+)(path)(=)(&quot;[^&]*&quot;)(&gt;)/g,
            '<span class="syntax-tag">$1</span><span class="syntax-attr">$2</span><span class="syntax-punct">$3</span><span class="syntax-string">$4</span><span class="syntax-tag">$5</span>')
        .replace(/(path)(=)(&quot;[^&]*&quot;)/g,
            '<span class="syntax-attr">$1</span><span class="syntax-punct">$2</span><span class="syntax-string">$3</span>')
}

export function highlightMarkdown(line: string): string {
    if (line.startsWith('## File:') || line.startsWith('# ')) {
        return `<span class="syntax-heading">${line}</span>`
    }
    if (line.startsWith('```')) {
        return `<span class="syntax-fence">${line}</span>`
    }
    return line
}

export function highlightJson(line: string): string {
    return line
        .replace(/(&quot;[^&]+&quot;)(\s*:)/g, '<span class="syntax-key">$1</span><span class="syntax-punct">$2</span>')
        .replace(/:(\s*)(&quot;[^&]*&quot;)/g, ':<span class="syntax-string">$1$2</span>')
        .replace(/:\s*(true|false|null|\d+)/g, ': <span class="syntax-value">$1</span>')
}

export function highlightPlain(line: string): string {
    if (line.startsWith('--- File:') && line.endsWith('---')) {
        return `<span class="syntax-separator">${line}</span>`
    }
    return line
}

export function applySyntaxHighlight(line: string, format: string): string {
    switch (format) {
        case 'xml':
            return highlightXml(line)
        case 'markdown':
            return highlightMarkdown(line)
        case 'json':
            return highlightJson(line)
        default:
            return highlightPlain(line)
    }
}

export function highlightLine(line: string, format: string, searchQuery?: string): string {
    let result = escapeHtml(line)

    // Apply syntax highlighting
    result = applySyntaxHighlight(result, format)

    // Apply search highlighting on top
    if (searchQuery) {
        const query = escapeHtml(searchQuery)
        const regex = new RegExp(`(${escapeRegex(query)})`, 'gi')
        result = result.replace(regex, '<mark class="bg-yellow-400/40 text-yellow-200 rounded px-0.5">$1</mark>')
    }

    return result
}
