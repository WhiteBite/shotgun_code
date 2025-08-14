#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Final cleanup for Shotgun App - fix syntax errors and add test scripts
"""

import json
import os
from pathlib import Path

ROOT = Path(__file__).resolve().parent
CHANGES = []

def write_text(p: Path, content: str):
    p.parent.mkdir(parents=True, exist_ok=True)
    p.write_text(content, encoding="utf-8")
    CHANGES.append(f"FIXED {p.as_posix()}")

def fix_filepanel_syntax():
    """Fix FilePanel.vue syntax error - remove orphaned code"""
    p = ROOT / "frontend" / "src" / "components" / "panels" / "FilePanel.vue"

    correct_content = """<template>
  <aside class="w-80 bg-gray-800/60 p-3 border-r border-gray-700 flex flex-col flex-shrink-0">
    <div class="flex-shrink-0 mb-2 flex items-center gap-2">
      <input
          v-model="contextStore.searchQuery"
          type="text"
          placeholder="Filter files..."
          class="w-full px-3 py-1.5 bg-gray-900 border border-gray-600 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
      <button @click="rescanFiles" class="p-2 rounded-md hover:bg-gray-700" title="Rescan Project Files">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-gray-400"><polyline points="23 4 23 10 17 10"></polyline><polyline points="1 20 1 14 7 14"></polyline><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path></svg>
      </button>
    </div>

    <div class="flex-grow bg-gray-900/50 rounded-md border border-gray-700 overflow-hidden min-h-0">
      <div v-if="contextStore.isLoading" class="p-4 text-center text-gray-400">
        Loading file tree...
      </div>
      <FileTree v-else :nodes="visibleNodes" />
    </div>

    <div class="flex-shrink-0 mt-2 space-y-3 pt-2 border-t border-gray-700/50">
      <div>
        <h3 class="font-semibold text-xs mb-2 text-gray-400">Git</h3>
        <button @click="gitStore.showHistory" class="w-full text-left p-2 text-sm bg-gray-900/50 hover:bg-gray-700/80 rounded-md">
          Commit History
        </button>
      </div>
      <div>
        <h3 class="font-semibold text-xs mb-2 text-gray-400">Ignore Rules</h3>
        <div class="space-y-2 text-sm text-gray-300">
          <label class="flex items-center gap-2 cursor-pointer">
            <input type="checkbox" v-model="settingsStore.settings.useGitignore" @change="updateIgnoreRules" class="form-checkbox bg-gray-700 border-gray-500 rounded text-blue-500 focus:ring-blue-500/50">
            Use .gitignore
          </label>
          <label class="flex items-center gap-2 cursor-pointer">
            <input type="checkbox" v-model="settingsStore.settings.useCustomIgnore" @change="updateIgnoreRules" class="form-checkbox bg-gray-700 border-gray-500 rounded text-blue-500 focus:ring-blue-500/50">
            Custom Rules
            <button @click="uiStore.openDrawer('ignore')" class="text-xs text-blue-400 hover:underline">(Edit)</button>
          </label>
        </div>
      </div>
      <ContextSummary />
    </div>
    <CommitHistoryModal />
  </aside>
</template>

<script setup lang="ts">
import { useContextStore } from '@/stores/context.store';
import { useSettingsStore } from '@/stores/settings.store';
import { useUiStore } from '@/stores/ui.store';
import { useGitStore } from '@/stores/git.store';
import { useVisibleNodes } from '@/composables/useVisibleNodes';
import FileTree from '@/components/workspace/FileTree.vue';
import ContextSummary from '@/components/workspace/ContextSummary.vue';
import CommitHistoryModal from '@/components/modals/CommitHistoryModal.vue';
import { useTreeStateStore } from '@/stores/tree-state.store';

const contextStore = useContextStore();
const settingsStore = useSettingsStore();
const uiStore = useUiStore();
const gitStore = useGitStore();
const treeStateStore = useTreeStateStore();
const { visibleNodes } = useVisibleNodes();

async function updateIgnoreRules() {
  await settingsStore.saveIgnoreSettings();
}

function rescanFiles() {
  contextStore.fetchFileTree();
}
</script>"""

    write_text(p, correct_content)

def update_package_json_scripts():
    """Add comprehensive test and validation scripts to package.json"""
    p = ROOT / "frontend" / "package.json"

    try:
        data = json.loads(p.read_text(encoding="utf-8"))

        # Update scripts section with validation and test commands
        data["scripts"] = {
            "dev": "vite",
            "build": "run-p type-check build-only",
            "build-only": "vite build",
            "preview": "vite preview",
            "test": "run-s test:unit test:lint test:type",
            "test:unit": "vitest run",
            "test:unit:watch": "vitest",
            "test:lint": "eslint . --ext .vue,.js,.jsx,.cjs,.mjs,.ts,.tsx,.cts,.mts --ignore-path .gitignore",
            "test:type": "vue-tsc --noEmit",
            "lint": "eslint . --ext .vue,.js,.jsx,.cjs,.mjs,.ts,.tsx,.cts,.mts --fix --ignore-path .gitignore",
            "type-check": "vue-tsc --noEmit",
            "validate": "run-s validate:*",
            "validate:syntax": "node scripts/validate-syntax.js",
            "validate:vue": "vue-tsc --noEmit --skipLibCheck",
            "validate:css": "stylelint \"src/**/*.{css,scss,vue}\"",
            "format": "prettier --write \"src/**/*.{js,ts,vue,css,json,md}\"",
            "format:check": "prettier --check \"src/**/*.{js,ts,vue,css,json,md}\"",
            "clean": "rimraf dist node_modules/.vite",
            "analyze": "vite build --mode analyze",
            "check-all": "run-s clean lint type-check test build"
        }

        # Add missing devDependencies for validation
        if "devDependencies" not in data:
            data["devDependencies"] = {}

        deps_to_add = {
            "vue-tsc": "^2.0.0",
            "npm-run-all": "^4.1.5",
            "rimraf": "^5.0.5",
            "prettier": "^3.2.5",
            "@vue/test-utils": "^2.4.5",
            "jsdom": "^24.0.0",
            "stylelint": "^16.2.1",
            "stylelint-config-standard": "^36.0.0",
            "stylelint-config-standard-vue": "^1.0.0",
            "rollup-plugin-visualizer": "^5.12.0"
        }

        for dep, version in deps_to_add.items():
            if dep not in data["devDependencies"]:
                data["devDependencies"][dep] = version

        p.write_text(json.dumps(data, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
        CHANGES.append(f"UPDATED {p.as_posix()} - added test/validation scripts")

    except Exception as e:
        CHANGES.append(f"ERROR updating {p.as_posix()}: {e}")

def create_validation_script():
    """Create a Node.js script to validate syntax of all JS/TS/Vue files"""
    scripts_dir = ROOT / "frontend" / "scripts"
    scripts_dir.mkdir(exist_ok=True)

    validation_script = scripts_dir / "validate-syntax.js"

    content = """#!/usr/bin/env node
/**
 * Syntax validation script for all JavaScript, TypeScript, and Vue files
 */

const fs = require('fs');
const path = require('path');
const { parse } = require('@babel/parser');
const glob = require('glob');

const errors = [];
let filesChecked = 0;

// Patterns to check
const patterns = [
  'src/**/*.js',
  'src/**/*.ts',
  'src/**/*.jsx',
  'src/**/*.tsx',
  'src/**/*.vue'
];

function checkFile(filePath) {
  filesChecked++;
  try {
    const content = fs.readFileSync(filePath, 'utf8');

    // For Vue files, extract script content
    if (filePath.endsWith('.vue')) {
      const scriptMatch = content.match(/<script[^>]*>(.*?)<\/script>/s);
      if (scriptMatch) {
        const scriptContent = scriptMatch[1];
        validateJS(scriptContent, filePath);
      }
    } else {
      validateJS(content, filePath);
    }
  } catch (error) {
    errors.push({
      file: filePath,
      error: error.message
    });
  }
}

function validateJS(content, filePath) {
  try {
    parse(content, {
      sourceType: 'module',
      plugins: [
        'typescript',
        'jsx',
        'decorators-legacy',
        'dynamicImport',
        'importMeta'
      ]
    });
  } catch (error) {
    errors.push({
      file: filePath,
      line: error.loc ? error.loc.line : 'unknown',
      column: error.loc ? error.loc.column : 'unknown',
      error: error.message
    });
  }
}

console.log('🔍 Validating syntax of all source files...');

patterns.forEach(pattern => {
  const files = glob.sync(pattern, {
    cwd: path.join(__dirname, '..'),
    absolute: true
  });
  files.forEach(checkFile);
});

console.log(`\\n✅ Checked ${filesChecked} files`);

if (errors.length > 0) {
  console.error(`\\n❌ Found ${errors.length} syntax errors:\\n`);
  errors.forEach(err => {
    console.error(`  ${err.file}`);
    if (err.line) {
      console.error(`    Line ${err.line}:${err.column} - ${err.error}`);
    } else {
      console.error(`    ${err.error}`);
    }
  });
  process.exit(1);
} else {
  console.log('✨ All files have valid syntax!');
}
"""

    write_text(validation_script, content)

    # Also need to add glob and @babel/parser to package.json
    p = ROOT / "frontend" / "package.json"
    try:
        data = json.loads(p.read_text(encoding="utf-8"))
        if "devDependencies" not in data:
            data["devDependencies"] = {}

        data["devDependencies"]["glob"] = "^10.3.10"
        data["devDependencies"]["@babel/parser"] = "^7.24.0"

        p.write_text(json.dumps(data, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    except Exception as e:
        CHANGES.append(f"ERROR adding babel/glob deps: {e}")

def create_prettier_config():
    """Create Prettier configuration"""
    p = ROOT / "frontend" / ".prettierrc.json"

    config = {
        "semi": true,
        "singleQuote": true,
        "tabWidth": 2,
        "trailingComma": "es5",
        "printWidth": 100,
        "arrowParens": "avoid",
        "vueIndentScriptAndStyle": false,
        "endOfLine": "auto"
    }

    write_text(p, json.dumps(config, indent=2) + "\n")

def create_stylelint_config():
    """Create Stylelint configuration"""
    p = ROOT / "frontend" / ".stylelintrc.json"

    config = {
        "extends": [
            "stylelint-config-standard",
            "stylelint-config-standard-vue"
        ],
        "rules": {
            "selector-class-pattern": None,
            "selector-id-pattern": None,
            "no-descending-specificity": None,
            "at-rule-no-unknown": [
                True,
                {
                    "ignoreAtRules": ["tailwind", "apply", "variants", "responsive", "screen"]
                }
            ]
        }
    }

    write_text(p, json.dumps(config, indent=2) + "\n")

def create_vitest_config():
    """Create basic test file example"""
    test_dir = ROOT / "frontend" / "src" / "__tests__"
    test_dir.mkdir(exist_ok=True)

    example_test = test_dir / "example.spec.ts"

    content = """import { describe, it, expect } from 'vitest';
import { mount } from '@vue/test-utils';

// Example test - remove or modify as needed
describe('Example Test Suite', () => {
  it('should pass basic assertion', () => {
    expect(1 + 1).toBe(2);
  });

  it('should validate environment', () => {
    expect(typeof window).toBe('object');
    expect(typeof document).toBe('object');
  });
});

// Example component test (when you have components to test)
/*
import MyComponent from '@/components/MyComponent.vue';

describe('MyComponent', () => {
  it('renders properly', () => {
    const wrapper = mount(MyComponent, {
      props: {
        msg: 'Hello Vitest'
      }
    });
    expect(wrapper.text()).toContain('Hello Vitest');
  });
});
*/
"""

    write_text(example_test, content)

def main():
    print("🔧 Applying final fixes and adding test infrastructure...\n")

    # Fix syntax errors
    fix_filepanel_syntax()

    # Add test/validation infrastructure
    update_package_json_scripts()
    create_validation_script()
    create_prettier_config()
    create_stylelint_config()
    create_vitest_config()

    print("📝 Changes made:")
    for change in CHANGES:
        print(f"  ✓ {change}")

    