import tailwindcss from '@tailwindcss/vite';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

const portalBasePlaceholder = '/__LETITCALL_BASE_PATH__';

export default defineConfig(({ command }) => ({
	plugins: [
		tailwindcss(),
		sveltekit({
			paths: {
				base: command === 'build' ? portalBasePlaceholder : ''
			},
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) => filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			adapter: adapter({ fallback: 'index.html' })
		})
	],
	server: {
		proxy: {
			'^/(api|content)(?:/|$)': {
				target: 'http://127.0.0.1:41784',
				changeOrigin: false
			}
		}
	}
}));
