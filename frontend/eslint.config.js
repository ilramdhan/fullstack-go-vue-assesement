import pluginVue from 'eslint-plugin-vue'
import vueTsConfig from '@vue/eslint-config-typescript'
import prettier from 'eslint-config-prettier'

export default [
  {
    ignores: ['dist/**', 'node_modules/**', 'src/api/generated/**', 'coverage/**'],
  },
  ...pluginVue.configs['flat/recommended'],
  ...vueTsConfig(),
  prettier,
  {
    rules: {
      'vue/multi-word-component-names': 'off',
    },
  },
]
