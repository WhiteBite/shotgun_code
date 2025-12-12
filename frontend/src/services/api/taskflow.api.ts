/**
 * Task Protocol and Guardrails API
 * Handles task execution protocols and security guardrails
 */

import * as wails from '#wailsjs/go/main/App'
import type { domain } from '#wailsjs/go/models'
import { apiCall } from './base'

export const taskflowApi = {
    // Task Protocol
    executeTaskProtocol: (configPath: string): Promise<string> =>
        apiCall(
            () => wails.ExecuteTaskProtocol(configPath),
            'Failed to execute task protocol.',
            'taskflow'
        ),

    getTaskProtocolConfiguration: (projectPath: string, languages: string[]): Promise<string> =>
        apiCall(
            () => wails.GetTaskProtocolConfiguration(projectPath, languages),
            'Failed to get task protocol configuration.',
            'taskflow'
        ),

    // Guardrails
    validatePath: (path: string): Promise<domain.GuardrailViolation[]> =>
        apiCall(() => wails.ValidatePath(path), 'Failed to validate path.', 'taskflow'),

    getGuardrailPolicies: (): Promise<domain.GuardrailPolicy[]> =>
        apiCall(
            () => wails.GetGuardrailPolicies(),
            'Failed to get guardrail policies.',
            'taskflow'
        ),

    getBudgetPolicies: (): Promise<domain.BudgetPolicy[]> =>
        apiCall(
            // @ts-ignore
            () => wails.GetBudgetPolicies(),
            'Failed to get budget policies.',
            'taskflow'
        ),
}
