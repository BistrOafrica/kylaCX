import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import { defineConfig, globalIgnores } from 'eslint/config'

export default defineConfig([
  // Generated protobuf stubs and quarantined legacy code aren't part of
  // the maintained codebase. They generate noise that drowns out real
  // findings, so skip them entirely.
  globalIgnores([
    'dist',
    'storybook-static',
    'coverage',
    'playwright-report',
    'src/pb/**',
    'src/_legacy/**',
  ]),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
    ],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
  },
  // shadcn primitives export a component alongside its CVA variants by
  // convention. That's fine — but trips react-refresh's "only export
  // components" rule. Same applies to Storybook preview decorators.
  {
    files: ['src/components/ui/**/*.{ts,tsx}', '.storybook/**/*.{ts,tsx}'],
    rules: {
      'react-refresh/only-export-components': 'off',
    },
  },
])
