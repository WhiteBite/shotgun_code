/**
 * UI Formatting Service
 * Encapsulates all UI formatting logic
 */
export class UIFormattingService {
  /**
   * Formats file size in human-readable format
   * @param bytes - The size in bytes
   * @returns string - The formatted size
   */
  formatFileSize(bytes: number): string {
    const sizes = ['B', 'KB', 'MB', 'GB'];
    if (bytes === 0) return '0 B';
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
  }

  /**
   * Formats a date in a human-readable format
   * @param dateString - The date string to format
   * @returns string - The formatted date
   */
  formatDate(dateString: string): string {
    try {
      const date = new Date(dateString);
      const now = new Date();
      const diffInHours = (now.getTime() - date.getTime()) / (1000 * 60 * 60);

      if (diffInHours < 1) {
        return "Just now";
      } else if (diffInHours < 24) {
        const hours = Math.floor(diffInHours);
        return `${hours} hour${hours > 1 ? "s" : ""} ago`;
      } else if (diffInHours < 24 * 7) {
        const days = Math.floor(diffInHours / 24);
        return `${days} day${days > 1 ? "s" : ""} ago`;
      } else {
        return date.toLocaleDateString();
      }
    } catch {
      return "Unknown";
    }
  }

  /**
   * Normalizes a path for display
   * @param path - The path to normalize
   * @returns string - The normalized path
   */
  normalizePath(path: string): string {
    return path ? path.replace(/\\/g, '/') : path;
  }

  /**
   * Truncates a string to a maximum length
   * @param str - The string to truncate
   * @param maxLength - The maximum length
   * @returns string - The truncated string
   */
  truncateString(str: string, maxLength: number): string {
    if (!str || str.length <= maxLength) return str;
    return str.substring(0, maxLength - 3) + '...';
  }

  /**
   * Formats a number with commas
   * @param num - The number to format
   * @returns string - The formatted number
   */
  formatNumber(num: number): string {
    return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
  }
}