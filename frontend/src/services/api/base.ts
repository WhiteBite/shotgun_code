/**
 * Base utilities for API modules
 * Provides common error handling and logging
 */

/**
 * Wrapper for API calls with consistent error handling
 */
export async function apiCall<T>(
    fn: () => Promise<T>,
    errorMsg: string,
    logContext?: string
): Promise<T> {
    try {
        return await fn()
    } catch (error) {
        console.error(`[API${logContext ? `:${logContext}` : ''}] ${errorMsg}:`, error)
        throw new Error(errorMsg)
    }
}

/**
 * Wrapper for API calls that return a default value on error
 */
export async function apiCallWithDefault<T>(
    fn: () => Promise<T>,
    defaultValue: T,
    logContext?: string
): Promise<T> {
    try {
        return await fn()
    } catch (error) {
        if (logContext) {
            console.warn(`[API:${logContext}] Error (using default):`, error)
        }
        return defaultValue
    }
}

/**
 * Parse JSON response with error handling
 */
export function parseJsonResponse<T>(json: string, errorMsg: string): T {
    try {
        return JSON.parse(json)
    } catch {
        throw new Error(errorMsg)
    }
}
