import { branding, loadCachedBranding } from '$lib/stores/branding.svelte';
import { pageTheme } from '$lib/public-theme';
import type { Theme } from '$lib/stores/theme';

loadCachedBranding();

// Prevent theme flash on page load by applying theme before rendering
function applyThemeBeforeRender() {
	if (typeof window === 'undefined') return;

	const saved = localStorage.getItem('theme');
	const theme: Theme = saved === 'light' || saved === 'dark'
		? saved
		: window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';

	if (pageTheme(window.location.pathname, theme, branding.publicTheme) === 'dark') {
		document.documentElement.classList.add('dark');
	} else {
		document.documentElement.classList.remove('dark');
	}
}

applyThemeBeforeRender();
