// Prevent theme flash on page load by applying theme before rendering
function applyThemeBeforeRender() {
	if (typeof window === 'undefined') return;

	const saved = localStorage.getItem('theme');
	let theme = saved === 'light' || saved === 'dark' ? saved : null;

	// Fall back to system preference if no saved preference
	if (!theme) {
		theme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
	}

	if (theme === 'dark') {
		document.documentElement.classList.add('dark');
	} else {
		document.documentElement.classList.remove('dark');
	}
}

applyThemeBeforeRender();
