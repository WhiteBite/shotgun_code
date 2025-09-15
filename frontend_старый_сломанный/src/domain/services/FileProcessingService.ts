/**
 * File Processing Service
 * Encapsulates file processing logic for UI components
 */
export class FileProcessingService {
  /**
   * Gets the file extension from a file name
   * @param fileName - The file name
   * @returns string - The file extension
   */
  getFileExtension(fileName: string): string {
    if (!fileName) return '';
    return fileName.split('.').pop()?.toLowerCase() || '';
  }

  /**
   * Gets the file name without extension
   * @param fileName - The file name
   * @returns string - The file name without extension
   */
  getFileNameWithoutExtension(fileName: string): string {
    if (!fileName) return '';
    const parts = fileName.split('.');
    return parts.length > 1 ? parts.slice(0, -1).join('.') : fileName;
  }

  /**
   * Gets the relative path from a full path and root path
   * @param fullPath - The full path
   * @param rootPath - The root path
   * @returns string - The relative path
   */
  getRelativePath(fullPath: string, rootPath: string): string {
    if (!fullPath || !rootPath) return fullPath || '';
    
    // Normalize paths
    const normalizedFull = fullPath.replace(/\\/g, '/');
    const normalizedRoot = rootPath.replace(/\\/g, '/');
    
    // Remove root path prefix
    if (normalizedFull.startsWith(normalizedRoot)) {
      const relative = normalizedFull.substring(normalizedRoot.length);
      // Remove leading slash if present
      return relative.startsWith('/') ? relative.substring(1) : relative;
    }
    
    return fullPath;
  }

  /**
   * Estimates token count for a file content
   * @param content - The file content
   * @returns number - The estimated token count
   */
  estimateTokenCount(content: string): number {
    if (!content) return 0;
    
    // Simple heuristic: 1 token ≈ 4 characters
    // This is a rough estimate and should be replaced with a more accurate method
    return Math.ceil(content.length / 4);
  }

  /**
   * Counts lines in file content
   * @param content - The file content
   * @returns number - The line count
   */
  countLines(content: string): number {
    if (!content) return 0;
    return content.split('\n').length;
  }

  /**
   * Gets a preview of file content
   * @param content - The file content
   * @param maxLength - The maximum length of preview
   * @returns string - The content preview
   */
  getFilePreview(content: string, maxLength: number = 200): string {
    if (!content) return '';
    if (content.length <= maxLength) return content;
    return content.substring(0, maxLength) + '...';
  }
}