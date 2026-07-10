import type { Config } from 'tailwindcss';

export default {
	darkMode: 'class',
	theme: {
		extend: {
			colors: {
				// Override with CSS variables for theming
				foreground: 'rgb(var(--color-foreground) / <alpha-value>)',
				background: 'rgb(var(--color-background) / <alpha-value>)',
				border: 'rgb(var(--color-border) / <alpha-value>)',
				'muted-foreground': 'rgb(var(--color-muted-foreground) / <alpha-value>)',
				'muted-background': 'rgb(var(--color-muted-background) / <alpha-value>)'
			}
		}
	}
} satisfies Config;
