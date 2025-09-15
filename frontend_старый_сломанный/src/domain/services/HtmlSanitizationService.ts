/**
 * HTML Sanitization Service
 * 
 * Domain service that provides comprehensive HTML sanitization following DDD principles.
 * Implements DOMPurify-like functionality with centralized configuration.
 */

import { APP_CONFIG } from '@/config/app-config';

export interface SanitizationConfig {
  allowedTags: string[];
  allowedAttributes: string[];
  forbiddenTags: string[];
  forbiddenAttributes: string[];
  allowedSchemes: string[];
  allowDataAttributes: boolean;
  stripScripts: boolean;
  stripStyles: boolean;
  stripComments: boolean;
  normalizeWhitespace: boolean;
  maxLength: number;
}

export interface SanitizationResult {
  sanitized: string;
  removed: {
    tags: string[];
    attributes: string[];
    scripts: number;
    styles: number;
    comments: number;
  };
  warnings: string[];
  isClean: boolean;
  originalLength: number;
  finalLength: number;
}

export interface SanitizationOptions {
  config?: Partial<SanitizationConfig>;
  preserveWhitespace?: boolean;
  returnDiagnostics?: boolean;
  strictMode?: boolean;
}

export class HtmlSanitizationService {
  private defaultConfig: SanitizationConfig;

  constructor() {
    this.defaultConfig = {
      allowedTags: APP_CONFIG.security.sanitization.ALLOWED_HTML_TAGS,
      allowedAttributes: APP_CONFIG.security.sanitization.ALLOWED_ATTRIBUTES,
      forbiddenTags: ['script', 'object', 'embed', 'iframe', 'frame', 'frameset', 'noframes', 'meta', 'link', 'style', 'form', 'input', 'textarea', 'button', 'select', 'option'],
      forbiddenAttributes: ['onload', 'onerror', 'onclick', 'onmouseover', 'onmouseout', 'onkeydown', 'onkeyup', 'onfocus', 'onblur', 'onsubmit', 'onreset', 'onchange', 'javascript:', 'vbscript:', 'data:'],
      allowedSchemes: ['http', 'https', 'mailto', 'tel'],
      allowDataAttributes: false,
      stripScripts: APP_CONFIG.security.sanitization.STRIP_SCRIPTS,
      stripStyles: true,
      stripComments: true,
      normalizeWhitespace: true,
      maxLength: 100000 // 100KB limit
    };
  }

  /**
   * Sanitize HTML content
   */
  sanitize(html: string, options: SanitizationOptions = {}): SanitizationResult {
    const config: SanitizationConfig = { ...this.defaultConfig, ...options.config };
    const originalLength = html.length;
    
    const result: SanitizationResult = {
      sanitized: '',
      removed: {
        tags: [],
        attributes: [],
        scripts: 0,
        styles: 0,
        comments: 0
      },
      warnings: [],
      isClean: true,
      originalLength,
      finalLength: 0
    };

    // Check length limit
    if (html.length > config.maxLength) {
      result.warnings.push(`Content exceeds maximum length of ${config.maxLength} characters`);
      html = html.substring(0, config.maxLength);
    }

    // Create DOM parser (safe in browser environment)
    const parser = new DOMParser();
    let doc: Document;
    
    try {
      doc = parser.parseFromString(html, 'text/html');
    } catch (error) {
      result.warnings.push(`Failed to parse HTML: ${error}`);
      result.sanitized = this.escapeHtml(html);
      result.finalLength = result.sanitized.length;
      result.isClean = false;
      return result;
    }

    // Process the document
    this.processNode(doc.body || doc.documentElement, config, result);

    // Extract sanitized content
    if (doc.body) {
      result.sanitized = doc.body.innerHTML;
    } else {
      result.sanitized = doc.documentElement.innerHTML;
    }

    // Post-processing
    if (config.normalizeWhitespace && !options.preserveWhitespace) {
      result.sanitized = this.normalizeWhitespace(result.sanitized);
    }

    // Remove empty attributes
    result.sanitized = this.removeEmptyAttributes(result.sanitized);

    result.finalLength = result.sanitized.length;
    result.isClean = result.removed.tags.length === 0 && 
                     result.removed.attributes.length === 0 && 
                     result.removed.scripts === 0;

    return result;
  }

  /**
   * Quick sanitization for simple text with minimal HTML
   */
  sanitizeSimple(html: string): string {
    const result = this.sanitize(html, {
      config: {
        allowedTags: ['strong', 'em', 'code', 'br'],
        allowedAttributes: [],
        stripScripts: true,
        stripStyles: true,
        stripComments: true
      }
    });
    return result.sanitized;
  }

  /**
   * Escape HTML entities
   */
  escapeHtml(text: string): string {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  /**
   * Check if HTML is safe (no dangerous content)
   */
  isSafe(html: string): boolean {
    const result = this.sanitize(html, { returnDiagnostics: true });
    return result.isClean && result.warnings.length === 0;
  }

  /**
   * Validate HTML against security policies
   */
  validateHtml(html: string): { isValid: boolean; violations: string[]; suggestions: string[] } {
    const violations: string[] = [];
    const suggestions: string[] = [];

    // Check for script tags
    if (/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi.test(html)) {
      violations.push('Script tags are not allowed');
      suggestions.push('Remove all <script> tags and use proper event handling');
    }

    // Check for javascript: protocol
    if (/javascript:/gi.test(html)) {
      violations.push('JavaScript protocol in URLs is not allowed');
      suggestions.push('Use http/https URLs instead of javascript: protocol');
    }

    // Check for event handlers
    const eventHandlers = /\s+on\w+\s*=/gi;
    if (eventHandlers.test(html)) {
      violations.push('Inline event handlers are not allowed');
      suggestions.push('Use proper event listeners instead of inline handlers');
    }

    // Check for style attributes with javascript
    if (/style\s*=.*javascript/gi.test(html)) {
      violations.push('JavaScript in style attributes is not allowed');
      suggestions.push('Use CSS classes instead of inline styles with JavaScript');
    }

    // Check for data URIs with script content
    if (/data:.*script/gi.test(html)) {
      violations.push('Data URIs with script content are not allowed');
      suggestions.push('Use proper image/media URLs instead of data URIs with executable content');
    }

    return {
      isValid: violations.length === 0,
      violations,
      suggestions
    };
  }

  /**
   * Get sanitization statistics
   */
  getStatistics(html: string): {
    totalTags: number;
    allowedTags: number;
    removedTags: number;
    totalAttributes: number;
    allowedAttributes: number;
    removedAttributes: number;
    hasScripts: boolean;
    hasStyles: boolean;
    hasEventHandlers: boolean;
  } {
    const result = this.sanitize(html, { returnDiagnostics: true });
    
    const tagMatches = html.match(/<[^>]+>/g) || [];
    const attributeMatches = html.match(/\s+\w+\s*=/g) || [];
    
    return {
      totalTags: tagMatches.length,
      allowedTags: tagMatches.length - result.removed.tags.length,
      removedTags: result.removed.tags.length,
      totalAttributes: attributeMatches.length,
      allowedAttributes: attributeMatches.length - result.removed.attributes.length,
      removedAttributes: result.removed.attributes.length,
      hasScripts: result.removed.scripts > 0,
      hasStyles: result.removed.styles > 0,
      hasEventHandlers: /\s+on\w+\s*=/gi.test(html)
    };
  }

  // Private helper methods

  private processNode(node: Node, config: SanitizationConfig, result: SanitizationResult): void {
    if (node.nodeType === Node.TEXT_NODE) {
      // Text nodes are generally safe, just return
      return;
    }

    if (node.nodeType === Node.COMMENT_NODE) {
      if (config.stripComments) {
        result.removed.comments++;
        node.parentNode?.removeChild(node);
        result.isClean = false;
      }
      return;
    }

    if (node.nodeType === Node.ELEMENT_NODE) {
      const element = node as Element;
      const tagName = element.tagName.toLowerCase();

      // Check if tag is forbidden
      if (config.forbiddenTags.includes(tagName)) {
        result.removed.tags.push(tagName);
        result.isClean = false;
        
        if (tagName === 'script') {
          result.removed.scripts++;
        } else if (tagName === 'style') {
          result.removed.styles++;
        }
        
        element.parentNode?.removeChild(element);
        return;
      }

      // Check if tag is allowed
      if (!config.allowedTags.includes(tagName)) {
        result.removed.tags.push(tagName);
        result.isClean = false;
        element.parentNode?.removeChild(element);
        return;
      }

      // Process attributes
      const attributesToRemove: string[] = [];
      
      for (let i = 0; i < element.attributes.length; i++) {
        const attr = element.attributes[i];
        const attrName = attr.name.toLowerCase();
        const attrValue = attr.value.toLowerCase();

        // Check forbidden attributes
        if (config.forbiddenAttributes.some(forbidden => 
          attrName.includes(forbidden) || attrValue.includes(forbidden)
        )) {
          attributesToRemove.push(attrName);
          continue;
        }

        // Check allowed attributes
        if (!config.allowedAttributes.includes(attrName)) {
          // Special case for data attributes
          if (attrName.startsWith('data-') && config.allowDataAttributes) {
            continue;
          }
          attributesToRemove.push(attrName);
          continue;
        }

        // Validate URL schemes in href and src attributes
        if ((attrName === 'href' || attrName === 'src') && attr.value) {
          if (!this.isValidUrl(attr.value, config.allowedSchemes)) {
            attributesToRemove.push(attrName);
          }
        }
      }

      // Remove forbidden attributes
      attributesToRemove.forEach(attrName => {
        element.removeAttribute(attrName);
        result.removed.attributes.push(attrName);
        result.isClean = false;
      });

      // Process child nodes
      const childNodes = Array.from(element.childNodes);
      childNodes.forEach(child => this.processNode(child, config, result));
    }
  }

  private isValidUrl(url: string, allowedSchemes: string[]): boolean {
    try {
      const parsedUrl = new URL(url, window.location.origin);
      return allowedSchemes.includes(parsedUrl.protocol.replace(':', ''));
    } catch {
      // If URL parsing fails, it's likely a relative URL or invalid
      // Allow relative URLs but be cautious
      return !url.toLowerCase().includes('javascript:') && 
             !url.toLowerCase().includes('vbscript:') &&
             !url.toLowerCase().includes('data:');
    }
  }

  private normalizeWhitespace(html: string): string {
    return html
      .replace(/\s+/g, ' ')  // Replace multiple whitespace with single space
      .replace(/>\s+</g, '><')  // Remove whitespace between tags
      .trim();
  }

  private removeEmptyAttributes(html: string): string {
    return html.replace(/\s+\w+\s*=\s*['"]\s*['"]/g, '');
  }
}

// Default instance for dependency injection
export const defaultHtmlSanitizationService = new HtmlSanitizationService();