import type { Config } from 'tailwindcss';

export default {
	darkMode: 'class',
	theme: {
			extend: {
			colors: {
				// Override with CSS variables for theming
				foreground: 'rgb(var(--color-foreground) / <alpha-value>)',
				background: 'rgb(var(--color-background) / <alpha-value>)',
				text: 'rgb(var(--color-text) / <alpha-value>)',
				'contrast-text': 'rgb(var(--color-contrast-text) / <alpha-value>)',
				border: 'rgb(var(--color-border) / <alpha-value>)',
				'muted-foreground': 'rgb(var(--color-muted-foreground) / <alpha-value>)',
				'muted-background': 'rgb(var(--color-muted-background) / <alpha-value>)',
				primary: 'rgb(var(--color-primary) / <alpha-value>)',
				secondary: 'rgb(var(--color-secondary) / <alpha-value>)',
				'secondary-hover': 'rgb(var(--color-secondary-hover) / <alpha-value>)'
			}
		}
	}
} satisfies Config;
