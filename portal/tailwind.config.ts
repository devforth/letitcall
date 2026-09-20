import type { Config } from 'tailwindcss';

export default {
	darkMode: 'class',
	theme: {
			extend: {
			colors: {
				// Override with CSS variables for theming
				background: 'rgb(var(--color-background) / <alpha-value>)',
				text: 'rgb(var(--color-text) / <alpha-value>)',
				'contrast-text': 'rgb(var(--color-contrast-text) / <alpha-value>)',
				border: 'rgb(var(--color-border) / <alpha-value>)',
				primary: 'rgb(var(--color-primary) / <alpha-value>)'
			}
		}
	}
} satisfies Config;
