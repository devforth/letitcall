import { writable } from 'svelte/store';

export type Theme = 'light' | 'dark';

function createThemeStore() {
	// Check for saved preference or system preference
	function getInitialTheme(): Theme {
		if (typeof window !== 'undefined') {
			const saved = localStorage.getItem('theme') as Theme | null;
			if (saved === 'light' || saved === 'dark') {
				return saved;
			}

			// Fall back to system preference
			if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
				return 'dark';
			}
		}

		return 'light';
	}

	const { subscribe, set } = writable<Theme>(getInitialTheme());

	let currentTheme: Theme = getInitialTheme();
	subscribe((theme) => {
		currentTheme = theme;
	});

	return {
		subscribe,
		setTheme: (theme: Theme) => {
			set(theme);
			if (typeof window !== 'undefined') {
				localStorage.setItem('theme', theme);
				// Update DOM for Tailwind dark mode
				if (theme === 'dark') {
					document.documentElement.classList.add('dark');
				} else {
					document.documentElement.classList.remove('dark');
				}
			}
		},
		toggle: () => {
			const newTheme = currentTheme === 'light' ? 'dark' : 'light';
			return newTheme;
		}
	};
}

export const theme = createThemeStore();
