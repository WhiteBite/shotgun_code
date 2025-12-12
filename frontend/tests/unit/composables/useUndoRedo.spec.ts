import { useUndoRedo } from '@/composables/useUndoRedo'
import { describe, expect, it } from 'vitest'

describe('useUndoRedo', () => {
    it('pushState adds state to history', () => {
        const { pushState, historyLength } = useUndoRedo<Set<string>>(
            new Set(),
            (s) => JSON.stringify([...s]),
            (str) => new Set(JSON.parse(str))
        )

        expect(historyLength.value).toBe(1) // Initial state

        pushState(new Set(['file1.ts']))
        expect(historyLength.value).toBe(2)

        pushState(new Set(['file1.ts', 'file2.ts']))
        expect(historyLength.value).toBe(3)
    })

    it('undo returns previous state', () => {
        const { pushState, undo, getCurrentState } = useUndoRedo<string[]>(
            [],
            JSON.stringify,
            JSON.parse
        )

        pushState(['a'])
        pushState(['a', 'b'])
        pushState(['a', 'b', 'c'])

        const prevState = undo()
        expect(prevState).toEqual(['a', 'b'])

        const prevState2 = undo()
        expect(prevState2).toEqual(['a'])
    })

    it('redo returns next state', () => {
        const { pushState, undo, redo } = useUndoRedo<string[]>(
            [],
            JSON.stringify,
            JSON.parse
        )

        pushState(['a'])
        pushState(['a', 'b'])

        undo() // Go back to ['a']

        const nextState = redo()
        expect(nextState).toEqual(['a', 'b'])
    })

    it('canUndo returns false at start', () => {
        const { canUndo } = useUndoRedo<string[]>([])

        expect(canUndo.value).toBe(false)
    })

    it('canRedo returns false at end', () => {
        const { pushState, canRedo } = useUndoRedo<string[]>([])

        pushState(['a'])
        expect(canRedo.value).toBe(false)
    })

    it('history is limited to 50 states', () => {
        const { pushState, historyLength } = useUndoRedo<number>(0)

        for (let i = 1; i <= 60; i++) {
            pushState(i)
        }

        expect(historyLength.value).toBe(50)
    })
})
