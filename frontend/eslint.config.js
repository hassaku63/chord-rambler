// @ts-check
import js from '@eslint/js';
import globals from 'globals';
import tseslint from 'typescript-eslint';
import reactHooks from 'eslint-plugin-react-hooks';
import testingLibrary from 'eslint-plugin-testing-library';
import vitest from '@vitest/eslint-plugin';
import prettier from 'eslint-config-prettier';
import { readdirSync } from 'node:fs';
import { join } from 'node:path';

// Architecture boundaries as lint rules (ADR-0008). Features are vertical
// slices that must not import each other; anything shared lives in
// src/shared. Components (*.tsx) stay free of side effects: no direct API
// calls, no browser globals; hooks own those.
const featureDirs = readdirSync(join(import.meta.dirname, 'src/features'), {
  withFileTypes: true,
})
  .filter((d) => d.isDirectory())
  .map((d) => d.name);

const featureBoundaryRules = featureDirs.map((name) => ({
  files: [`src/features/${name}/**/*.{ts,tsx}`],
  rules: {
    '@typescript-eslint/no-restricted-imports': [
      'error',
      {
        patterns: [
          {
            group: ['@/features/*', `!@/features/${name}`, `!@/features/${name}/**`],
            message: `features must not import each other; move shared code to src/shared (from ${name})`,
            allowTypeImports: true,
          },
          {
            group: ['../*', '!../*.css'],
            message:
              'use ./ inside a feature and @/shared for shared code; never reach into sibling features',
            allowTypeImports: true,
          },
        ],
      },
    ],
  },
}));

const componentPurityRules = {
  files: ['src/**/*.tsx'],
  ignores: ['src/**/*.test.tsx', 'src/main.tsx'],
  rules: {
    '@typescript-eslint/no-restricted-imports': [
      'error',
      {
        patterns: [
          {
            group: ['@/shared/api', '@/shared/api/*'],
            message:
              'components do not call the API; put the call in a hook (use*.ts) and pass results as props',
            allowTypeImports: true,
          },
        ],
      },
    ],
    'no-restricted-globals': [
      'error',
      ...[
        'fetch',
        'AudioContext',
        'localStorage',
        'sessionStorage',
        'indexedDB',
        'setTimeout',
        'setInterval',
      ].map((name) => ({
        name,
        message: `${name} is a side effect; keep it in a hook, not in a component`,
      })),
    ],
  },
};

const testHygieneRules = {
  files: ['src/**/*.test.{ts,tsx}'],
  rules: {
    'no-restricted-syntax': [
      'error',
      {
        selector: 'CallExpression[callee.property.name=/^toMatch(Inline)?Snapshot$/]',
        message:
          'snapshot tests are banned (docs/dev/workflow.md); assert on roles, text and data instead',
      },
    ],
  },
};

export default tseslint.config(
  {
    ignores: [
      'dist',
      'coverage',
      'node_modules',
      'test-results',
      'playwright-report',
      'src/**/*.gen.ts',
    ],
  },
  js.configs.recommended,
  ...tseslint.configs.strictTypeChecked,
  ...tseslint.configs.stylisticTypeChecked,
  {
    languageOptions: {
      parserOptions: {
        projectService: { allowDefaultProject: ['eslint.config.js'] },
        tsconfigRootDir: import.meta.dirname,
      },
      globals: { ...globals.browser },
    },
  },
  {
    files: ['**/*.{ts,tsx}'],
    plugins: { 'react-hooks': reactHooks },
    rules: {
      ...reactHooks.configs['recommended-latest'].rules,
      '@typescript-eslint/consistent-type-imports': 'error',
      '@typescript-eslint/no-floating-promises': 'error',
      // React handlers and test waits are written `() => setX(v)` and
      // `waitFor(() => expect(...))`; nothing reads their result.
      '@typescript-eslint/no-confusing-void-expression': ['error', { ignoreArrowShorthand: true }],
      // Numbers print the same everywhere.
      '@typescript-eslint/restrict-template-expressions': ['error', { allowNumber: true }],
      // Object shapes are written as `type` throughout; keep them so.
      '@typescript-eslint/consistent-type-definitions': ['error', 'type'],
      // Size and complexity budgets. Generous; a hit means "split".
      'max-lines': ['error', { max: 400, skipBlankLines: true, skipComments: true }],
      'max-lines-per-function': ['error', { max: 120, skipBlankLines: true, skipComments: true }],
      complexity: ['error', 15],
    },
  },
  {
    files: ['**/*.test.{ts,tsx}', 'src/test/**'],
    rules: { 'max-lines-per-function': 'off' },
  },
  // Testing Library's own rules: queries by role, awaited async utilities,
  // no container or node access.
  {
    files: ['src/**/*.test.{ts,tsx}'],
    ...testingLibrary.configs['flat/react'],
  },
  // Vitest's own rules: no identical titles, no focused or commented-out
  // tests, expect called the right way.
  {
    files: ['src/**/*.test.{ts,tsx}'],
    ...vitest.configs.recommended,
  },
  ...featureBoundaryRules,
  componentPurityRules,
  testHygieneRules,
  {
    files: ['*.config.{js,ts}'],
    languageOptions: { globals: { ...globals.node } },
  },
  {
    // Plain JS config files are not part of the TS project: no type-aware rules.
    files: ['**/*.js'],
    ...tseslint.configs.disableTypeChecked,
  },
  prettier,
);
