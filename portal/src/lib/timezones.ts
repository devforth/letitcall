export function getLocalTimezones(): { current: string; options: string[] } {
	const current = Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC';
	try {
		const timezoneIntl = Intl as typeof Intl & { supportedValuesOf?: (key: string) => string[] };
		const supported = timezoneIntl.supportedValuesOf?.('timeZone') ?? [];
		return { current, options: [...new Set([current, ...supported, 'UTC'])] };
	} catch {
		return { current, options: [...new Set([current, 'UTC'])] };
	}
}
