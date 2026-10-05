export function browserTimeFormatter(options: Intl.DateTimeFormatOptions): Intl.DateTimeFormat {
	return new Intl.DateTimeFormat(
		typeof navigator === 'undefined' ? undefined : navigator.languages,
		options
	);
}

export function browserLanguage(): string | undefined {
	return typeof navigator === 'undefined' ? undefined : navigator.language;
}
