import { setActivePinia, createPinia } from 'pinia';
import { useFileTreeStore } from './fileTreeStore';
import { describe, it, expect, beforeEach, vi } from 'vitest';

// Mock dependencies that are not under test.
vi.mock('../../wailsjs/go/main/App', () => ({
  ListFiles: vi.fn(),
  StartFileWatcher: vi.fn(),
  StopFileWatcher: vi.fn(),
}));
vi.mock('../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(),
}));
vi.mock('./notificationsStore.js', () => ({
  useNotificationsStore: () => ({ addLog: vi.fn() }),
}));
vi.mock('./settingsStore.js', () => ({
  useSettingsStore: () => ({ useGitignore: true, useCustomIgnore: true }),
}));
vi.mock('./projectStore.js', () => ({
  useProjectStore: () => ({ projectRoot: '' }),
}));


describe('fileTreeStore: applySelectionSet', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  // Helper to create a realistic node structure for tests
  const createNode = (relPath, isDir, children = []) => ({
    name: relPath.split('/').pop() || relPath,
    relPath,
    isDir,
    children,
    excluded: false,
    expanded: isDir && relPath === '.', // By default, only the root is expanded
    isGitignored: false,
    isCustomIgnored: false,
  });

  it('should correctly exclude nodes and expand ancestors of selected files', () => {
    const store = useFileTreeStore();

    // 1. Setup initial tree state
    store.tree = [
      createNode('.', true, [
        createNode('src', true, [
          createNode('src/main.js', false),
          createNode('src/components', true, [
            createNode('src/components/Button.vue', false),
          ]),
        ]),
        createNode('public', true, [
          createNode('public/index.html', false),
        ]),
        createNode('package.json', false),
      ]),
    ];

    const root = store.tree[0];
    const srcDir = root.children.find(c => c.name === 'src');
    const mainJs = srcDir.children.find(c => c.name === 'main.js');
    const componentsDir = srcDir.children.find(c => c.name === 'components');
    const buttonVue = componentsDir.children.find(c => c.name === 'Button.vue');
    const publicDir = root.children.find(c => c.name === 'public');
    const indexHtml = publicDir.children.find(c => c.name === 'index.html');
    const packageJson = root.children.find(c => c.name === 'package.json');

    // 2. Define the selection set
    const selectionSet = new Set(['src/components/Button.vue', 'public/index.html']);

    // 3. Run the action
    store.applySelectionSet(selectionSet);

    // 4. Assertions
    expect(root.excluded, 'Root should be included').toBe(false);
    expect(root.expanded, 'Root should be expanded').toBe(true);

    expect(srcDir.excluded, 'Ancestor directory "src" should be included').toBe(false);
    expect(srcDir.expanded, 'Ancestor directory "src" should be expanded').toBe(true);

    expect(componentsDir.excluded, 'Ancestor directory "components" should be included').toBe(false);
    expect(componentsDir.expanded, 'Ancestor directory "components" should be expanded').toBe(true);

    expect(publicDir.excluded, 'Ancestor directory "public" should be included').toBe(false);
    expect(publicDir.expanded, 'Ancestor directory "public" should be expanded').toBe(true);

    expect(buttonVue.excluded, 'Selected file "Button.vue" should be included').toBe(false);
    expect(indexHtml.excluded, 'Selected file "index.html" should be included').toBe(false);

    expect(mainJs.excluded, 'Unselected file "main.js" should be excluded').toBe(true);
    expect(packageJson.excluded, 'Unselected file "package.json" should be excluded').toBe(true);
  });

  it('should not change the tree state if the selection set is empty', () => {
    const store = useFileTreeStore();
    const initialTreeState = [
      createNode('.', true, [ createNode('file.txt', false) ])
    ];
    // Use deep copy to ensure the original state isn't mutated by reference
    store.tree = JSON.parse(JSON.stringify(initialTreeState));

    store.applySelectionSet(new Set());

    // The tree should remain identical to its initial state
    expect(store.tree).toEqual(initialTreeState);
  });
});