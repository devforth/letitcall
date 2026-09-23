import { callApi } from '$lib/api';
import type { Branding, BrandingTheme, ThemeColors } from '$lib/types';

const cacheKey = 'branding';

export const defaultBrandingTheme: BrandingTheme = {
	light: {
		primary: '#0284C7',
		text: '#646464',
		background: '#FFFFFF'
	},
	dark: {
		primary: '#0284C7',
		text: '#FFFFFF',
		background: '#646464'
	}
};

export const branding = $state<Branding>({
	name: 'Let It Call',
	logoPath: '',
	theme: structuredClone(defaultBrandingTheme),
	preset: 'custom'
});

const cssColorNames: (keyof ThemeColors)[] = [
	'primary',
	'text',
	'background'
];

function colorChannels(hex: string): string {
	return `${Number.parseInt(hex.slice(1, 3), 16)} ${Number.parseInt(hex.slice(3, 5), 16)} ${Number.parseInt(hex.slice(5, 7), 16)}`;
}

function applyTheme(theme: BrandingTheme) {
	for (const mode of ['light', 'dark'] as const) {
		for (const name of cssColorNames) {
			document.documentElement.style.setProperty(
				`--branding-${mode}-${name}`,
				colorChannels(theme[mode][name])
			);
		}
	}
}

export function applyBranding(value: Branding) {
	branding.name = value.name;
	branding.logoPath = value.logoPath ?? '';
	branding.theme = value.theme;
	branding.preset = value.preset ?? 'custom';
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
		if (!value.theme.light.background || value.theme.light.background === '#F5F5F0') {
			value.theme.light.background = defaultBrandingTheme.light.background;
		}
		value.theme.dark.background ||= defaultBrandingTheme.dark.background;
		applyBranding(value);
	}
}

export async function loadBranding(reportError = true): Promise<Branding> {
	const response = await callApi<{ branding: Branding }>('/api/branding', undefined, reportError);
	applyBranding(response.branding);
	return response.branding;
}
