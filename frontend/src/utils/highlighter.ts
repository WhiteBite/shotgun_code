import hljs from 'highlight.js'

// Import specific languages to reduce bundle size
import javascript from 'highlight.js/lib/languages/javascript'
import typescript from 'highlight.js/lib/languages/typescript'
import python from 'highlight.js/lib/languages/python'
import css from 'highlight.js/lib/languages/css'
import json from 'highlight.js/lib/languages/json'
import markdown from 'highlight.js/lib/languages/markdown'
import xml from 'highlight.js/lib/languages/xml'
import bash from 'highlight.js/lib/languages/bash'

// Register languages
hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('typescript', typescript)
hljs.registerLanguage('python', python)
hljs.registerLanguage('css', css)
hljs.registerLanguage('json', json)
hljs.registerLanguage('markdown', markdown)
hljs.registerLanguage('xml', xml)
hljs.registerLanguage('html', xml)
hljs.registerLanguage('bash', bash)

export function highlightCode(code: string, language?: string): string {
  try {
    if (language && hljs.getLanguage(language)) {
      return hljs.highlight(code, { language }).value
    } else {
      const result = hljs.highlightAuto(code)
      return result.value
    }
  } catch (error) {
    console.warn('Code highlighting failed:', error)
    return escapeHtml(code)
  }
}

export function detectLanguage(code: string): string {
  try {
    const result = hljs.highlightAuto(code)
    return result.language || 'text'
  } catch (error) {
    console.warn('Language detection failed:', error)
    return 'text'
  }
}

export function getLanguageDisplayName(language: string): string {
  const languageNames: Record<string, string> = {
    javascript: 'JavaScript',
    typescript: 'TypeScript',
    python: 'Python',
    css: 'CSS',
    json: 'JSON',
    markdown: 'Markdown',
    xml: 'XML',
    html: 'HTML',
    bash: 'Shell',
    text: 'Plain Text'
  }
  
  return languageNames[language] || language.charAt(0).toUpperCase() + language.slice(1)
}

function escapeHtml(text: string): string {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}

export default {
  highlightCode,
  detectLanguage,
  getLanguageDisplayName
}