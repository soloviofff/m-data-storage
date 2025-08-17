module.exports = {
	root: true,
	env: { node: true, es2022: true },
	parser: '@typescript-eslint/parser',
	parserOptions: { project: null, ecmaVersion: 'latest', sourceType: 'module' },
	plugins: ['@typescript-eslint', 'import', 'unused-imports'],
	extends: [
		'eslint:recommended',
		'plugin:@typescript-eslint/recommended',
		'plugin:import/recommended',
		'plugin:import/typescript',
		'prettier',
	],
	settings: { 'import/resolver': { node: { extensions: ['.js', '.ts'] } } },
	rules: {
		'no-console': 'off',
		'import/no-unresolved': 'off',
		'unused-imports/no-unused-imports': 'error',
		'@typescript-eslint/no-unused-vars': [
			'error',
			{ argsIgnorePattern: '^_', varsIgnorePattern: '^_' },
		],
	},
	ignorePatterns: ['dist', 'node_modules'],
};
