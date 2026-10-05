import type { Branding } from '$lib/types';
import type { Theme } from '$lib/stores/theme';

export function pageTheme(pathname: string, preferred: Theme, publicTheme: Branding['publicTheme']): Theme {
	return /\/(book|event)\/[^/]+\/?$/.test(pathname) && publicTheme !== 'both'
		? publicTheme
		: preferred;
}
