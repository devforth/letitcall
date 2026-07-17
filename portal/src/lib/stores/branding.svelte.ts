import { callApi } from '$lib/api';
import type { Branding, BrandingTheme, ThemeColors } from '$lib/types';

const cacheKey = 'branding';

export const defaultBrandingTheme: BrandingTheme = {
	light: {
		primary: '#00C950',
		primaryContrast: '#FFFFFF',
		foreground: '#FFFFFF',
		text: '#646464',
		background: '#F5F5F0',
		border: '#D8D8D8',
		shadow: '#000000'
	},
	dark: {
		primary: '#00C950',
		primaryContrast: '#FFFFFF',
		foreground: '#646464',
		text: '#FFFFFF',
		background: '#333333',
		border: '#787878',
		shadow: '#000000'
	}
};

export const branding = $state<Branding>({
	name: 'Let It Call',
	logoPath: '',
	theme: structuredClone(defaultBrandingTheme)
});

const cssColorNames: (keyof ThemeColors)[] = [
	'primary',
	'primaryContrast',
	'foreground',
	'text',
	'background',
	'border',
	'shadow'
];

function colorChannels(hex: string): string {
	return `${Number.parseInt(hex.slice(1, 3), 16)} ${Number.parseInt(hex.slice(3, 5), 16)} ${Number.parseInt(hex.slice(5, 7), 16)}`;
}

function applyTheme(theme: BrandingTheme) {
	for (const mode of ['light', 'dark'] as const) {
		for (const name of cssColorNames) {
			const cssName = name === 'primaryContrast' ? 'contrast-text' : name;
			document.documentElement.style.setProperty(
				`--branding-${mode}-${cssName}`,
				colorChannels(theme[mode][name])
			);
		}
	}
}

export function applyBranding(value: Branding) {
	branding.name = value.name;
	branding.logoPath = value.logoPath ?? '';
	branding.theme = value.theme;
	localStorage.setItem(cacheKey, JSON.stringify(value));
	applyTheme(value.theme);
}

export function loadCachedBranding() {
	const cached = localStorage.getItem(cacheKey);
	if (cached) {
		const value = JSON.parse(cached) as Branding;
		if (!value.theme) {
			value.theme = structuredClone(defaultBrandingTheme);
		}
		value.theme.light.border ||= defaultBrandingTheme.light.border;
		value.theme.dark.border ||= defaultBrandingTheme.dark.border;
		applyBranding(value);
	}
}

export async function loadBranding(reportError = true): Promise<Branding> {
	const response = await callApi<{ branding: Branding }>('/api/branding', undefined, reportError);
	applyBranding(response.branding);
	return response.branding;
}
